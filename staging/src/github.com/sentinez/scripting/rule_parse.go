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
	anyOf, rets, err := parseScript(source, MatchFunc)
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

	return Rule{Any: anyOf, Action: action}, nil
}

// ParseKey reads the branches of a rate limiter script of the form
// KeySource writes: as ParseMatch, each branch returning true.
func ParseKey(source string) ([][]Cond, error) {
	anyOf, rets, err := parseScript(source, KeyFunc)
	if err != nil {
		return nil, err
	}

	for _, ret := range rets {
		if _, ok := ret.(*ast.TrueExpr); !ok {
			return nil, lineErr(ret, "a key branch returns true: the "+
				"client IP")
		}
	}

	return anyOf, nil
}

// parseScript the branches of the function fn of source, and the
// expression each one returns.
func parseScript(source, fn string) ([][]Cond, []ast.Expr, error) {
	chunk, err := parse.Parse(strings.NewReader(source), fn)
	if err != nil {
		return nil, nil, err
	}

	p, body, err := parseFunc(chunk, fn)
	if err != nil {
		return nil, nil, err
	}

	for i, stmt := range body {
		last := i == len(body)-1
		if err := p.stmt(stmt, last); err != nil {
			return nil, nil, err
		}
	}

	if len(p.anyOf) == 0 {
		return nil, nil, fmt.Errorf("%w: %w", ErrNotDeclarative, errNoBranch)
	}

	return p.anyOf, p.rets, nil
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
	ctx   string
	anyOf [][]Cond
	rets  []ast.Expr
}

// stmt adds the branch of stmt, an if or, last, a return.
func (p *parser) stmt(stmt ast.Stmt, last bool) error {
	if ret, ok := stmt.(*ast.ReturnStmt); ok && last {
		expr, err := returned(ret)
		if err != nil {
			return err
		}

		return p.add([][]Cond{nil}, expr)
	}

	s, ok := stmt.(*ast.IfStmt)
	if !ok {
		return lineErr(stmt, "only if ... then return ... end statements")
	}

	if len(s.Else) > 0 || len(s.Then) != 1 {
		return lineErr(s, "an if holds one return, without else")
	}

	ret, ok := s.Then[0].(*ast.ReturnStmt)
	if !ok {
		return lineErr(s.Then[0], "an if holds one return")
	}

	expr, err := returned(ret)
	if err != nil {
		return err
	}

	branches, err := p.dnf(s.Condition, false)
	if err != nil {
		return err
	}

	return p.add(branches, expr)
}

func (p *parser) add(branches [][]Cond, ret ast.Expr) error {
	if len(p.anyOf)+len(branches) > _maxBranches {
		return lineErr(ret, "too many branches")
	}

	p.anyOf = append(p.anyOf, branches...)
	p.rets = append(p.rets, ret)

	return nil
}

func returned(ret *ast.ReturnStmt) (ast.Expr, error) {
	if len(ret.Exprs) != 1 {
		return nil, lineErr(ret, "return one value")
	}

	return ret.Exprs[0], nil
}

// dnf the branches of e, its negation when neg, in disjunctive normal
// form.
func (p *parser) dnf(e ast.Expr, neg bool) ([][]Cond, error) {
	switch x := e.(type) {
	case *ast.UnaryNotOpExpr:
		return p.dnf(x.Expr, !neg)
	case *ast.LogicalOpExpr:
		return p.logical(x, neg)
	}

	c, err := p.cond(e)
	if err != nil {
		return nil, err
	}

	if neg {
		if c, err = negate(c, e); err != nil {
			return nil, err
		}
	}

	return [][]Cond{{c}}, nil
}

// logical the branches of x, an and or an or, as De Morgan negates them
// when neg.
func (p *parser) logical(x *ast.LogicalOpExpr, neg bool) ([][]Cond, error) {
	lhs, err := p.dnf(x.Lhs, neg)
	if err != nil {
		return nil, err
	}

	rhs, err := p.dnf(x.Rhs, neg)
	if err != nil {
		return nil, err
	}

	and := (x.Operator == "and") != neg
	if !and {
		return append(lhs, rhs...), nil
	}

	if len(lhs)*len(rhs) > _maxBranches {
		return nil, lineErr(x, "too many branches")
	}

	out := make([][]Cond, 0, len(lhs)*len(rhs))
	for _, l := range lhs {
		for _, r := range rhs {
			out = append(out, append(append([]Cond{}, l...), r...))
		}
	}

	return out, nil
}

// negate c, the condition e reads: an equality only.
func negate(c Cond, e ast.Expr) (Cond, error) {
	switch c.Op {
	case OpEq:
		c.Op = OpNe
	case OpNe:
		c.Op = OpEq
	default:
		return Cond{}, lineErr(e, "not applies to == and ~= only")
	}

	return c, nil
}

func lineErr(n ast.PositionHolder, msg string) error {
	return fmt.Errorf("%w: line %d: %s", ErrNotDeclarative, n.Line(), msg)
}
