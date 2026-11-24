// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scripting

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/yuin/gopher-lua/ast"
	"github.com/yuin/gopher-lua/parse"
)

// ParseMatch reads the declarative Rule of a secrule script (see Rule):
// one function match(ctx) whose body is a list of
//
//	if <condition> then return <action> end
//
// every branch returning the same action, the last statement possibly an
// unconditional return. A condition combines with and, or, not and
// parentheses the comparisons MatchSource writes. Any other script returns
// an error wrapping ErrNotDeclarative, with the line at fault.
func ParseMatch(source string) (Rule, error) {
	match, rets, err := parseScript(source, MatchFunc)
	if err != nil {
		return Rule{}, err
	}

	var action Action
	for i, ret := range rets {
		a, err := parseAction(ret)
		if err != nil {
			return Rule{}, err
		}

		if i > 0 && !reflect.DeepEqual(a, action) {
			return Rule{}, lineErr(ret, "every branch must return the "+
				"same action: split it into one rule per action")
		}

		action = a
	}

	return Rule{Match: match, Action: action}, nil
}

// ParseKey reads the condition of a rate limiter script of the form
// KeySource writes: as ParseMatch, each branch returning true.
func ParseKey(source string) (Node, error) {
	match, rets, err := parseScript(source, KeyFunc)
	if err != nil {
		return Node{}, err
	}

	for _, ret := range rets {
		if _, ok := ret.(*ast.TrueExpr); !ok {
			return Node{}, lineErr(ret, "a key branch returns true: the "+
				"client IP")
		}
	}

	return match, nil
}

// parseScript the condition matching one of the branches of the function
// fn of source, and the expression each one returns.
func parseScript(source, fn string) (Node, []ast.Expr, error) {
	chunk, err := parse.Parse(strings.NewReader(source), fn)
	if err != nil {
		return Node{}, nil, err
	}

	p, body, err := parseFunc(chunk, fn)
	if err != nil {
		return Node{}, nil, err
	}

	for i, stmt := range body {
		last := i == len(body)-1
		if err := p.stmt(stmt, last); err != nil {
			return Node{}, nil, err
		}
	}

	if len(p.branches) == 0 {
		return Node{}, nil, fmt.Errorf("%w: %w", ErrNotDeclarative,
			errNoBranch)
	}

	match := Any(p.branches...)
	if match.count() > MaxConditions {
		return Node{}, nil, errTooMany
	}

	return match, p.rets, nil
}

// parseFunc the parser of the function fn, the only statement of chunk,
// and its body.
func parseFunc(chunk []ast.Stmt, fn string) (*parser, []ast.Stmt, error) {
	if len(chunk) != 1 {
		return nil, nil, fmt.Errorf("%w: the script is one function "+
			"%s(ctx) only", ErrNotDeclarative, fn)
	}

	def, ok := chunk[0].(*ast.FuncDefStmt)
	if !ok || def.Name.Receiver != nil {
		return nil, nil, lineErr(chunk[0], "function "+fn+"(ctx) expected")
	}

	name, ok := def.Name.Func.(*ast.IdentExpr)
	params := def.Func.ParList
	if !ok || name.Value != fn || len(params.Names) != 1 ||
		params.HasVargs {
		return nil, nil, lineErr(def, "function "+fn+"(ctx) expected")
	}

	return &parser{ctx: params.Names[0]}, def.Func.Stmts, nil
}

// parser reads the branches of a function, ctx the name of its parameter.
type parser struct {
	ctx      string
	branches []Node
	rets     []ast.Expr
}

// stmt adds the branch of stmt, an if or, last, a return.
func (p *parser) stmt(stmt ast.Stmt, last bool) error {
	branch, ret, err := p.branchOf(stmt, last)
	if err != nil {
		return err
	}

	p.branches = append(p.branches, branch)
	p.rets = append(p.rets, ret)

	return nil
}

// branchOf the condition of stmt and the expression it returns: a return,
// last, matches every request.
func (p *parser) branchOf(stmt ast.Stmt, last bool) (Node, ast.Expr, error) {
	if ret, ok := stmt.(*ast.ReturnStmt); ok && last {
		expr, err := returned(ret)

		return All(), expr, err
	}

	s, ok := stmt.(*ast.IfStmt)
	if !ok {
		return Node{}, nil, lineErr(stmt,
			"only if ... then return ... end statements")
	}

	if len(s.Else) > 0 || len(s.Then) != 1 {
		return Node{}, nil, lineErr(s, "an if holds one return, without else")
	}

	ret, ok := s.Then[0].(*ast.ReturnStmt)
	if !ok {
		return Node{}, nil, lineErr(s.Then[0], "an if holds one return")
	}

	expr, err := returned(ret)
	if err != nil {
		return Node{}, nil, err
	}

	branch, err := p.tree(s.Condition, false)

	return branch, expr, err
}

func returned(ret *ast.ReturnStmt) (ast.Expr, error) {
	if len(ret.Exprs) != 1 {
		return nil, lineErr(ret, "return one value")
	}

	return ret.Exprs[0], nil
}

// tree the node of e, its negation when neg: not is pushed down to the
// conditions, by De Morgan.
func (p *parser) tree(e ast.Expr, neg bool) (Node, error) {
	switch x := e.(type) {
	case *ast.UnaryNotOpExpr:
		return p.tree(x.Expr, !neg)
	case *ast.LogicalOpExpr:
		return p.logical(x, neg)
	}

	c, err := p.cond(e)
	if err != nil {
		return Node{}, err
	}

	if neg {
		if c, err = negate(c, e); err != nil {
			return Node{}, err
		}
	}

	return Leaf(c), nil
}

// logical the node of x, an and or an or, as De Morgan negates it when
// neg.
func (p *parser) logical(x *ast.LogicalOpExpr, neg bool) (Node, error) {
	lhs, err := p.tree(x.Lhs, neg)
	if err != nil {
		return Node{}, err
	}

	rhs, err := p.tree(x.Rhs, neg)
	if err != nil {
		return Node{}, err
	}

	if (x.Operator == "and") != neg {
		return All(lhs, rhs), nil
	}

	return Any(lhs, rhs), nil
}

// negate c, the condition e reads: an equality, an ip_in or a
// presence only.
func negate(c Cond, e ast.Expr) (Cond, error) {
	op, ok := _negations[c.Op]
	if !ok {
		return Cond{}, lineErr(e, "not applies to ==, ~=, ctx:tls(), "+
			"ctx:ip_in, ctx:has_header and ctx:has_query only")
	}

	c.Op = op

	return c, nil
}

func lineErr(n ast.PositionHolder, msg string) error {
	return fmt.Errorf("%w: line %d: %s", ErrNotDeclarative, n.Line(), msg)
}
