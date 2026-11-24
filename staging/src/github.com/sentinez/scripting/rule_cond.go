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
	"strconv"

	"github.com/yuin/gopher-lua/ast"
)

// cond the condition e, one of the comparisons MatchSource writes.
func (p *parser) cond(e ast.Expr) (Cond, error) {
	switch x := e.(type) {
	case *ast.RelationalOpExpr:
		return p.relational(x)
	case *ast.FuncCallExpr:
		return p.contains(x)
	}

	return Cond{}, lineErr(e, "unsupported condition")
}

// relational field == "v", field ~= "v", field:sub(...) == "v" or
// tonumber(field) <op> n.
func (p *parser) relational(x *ast.RelationalOpExpr) (Cond, error) {
	if c, ok := p.numeric(x); ok {
		return c, nil
	}

	v, ok := x.Rhs.(*ast.StringExpr)
	if !ok {
		return Cond{}, lineErr(x, "compare to a string literal")
	}

	c, ok := p.field(x.Lhs)
	switch {
	case ok && x.Operator == "==":
		c.Op = OpEq
	case ok && x.Operator == "~=":
		c.Op = OpNe
	case x.Operator == "==":
		return p.affix(x.Lhs, v.Value)
	default:
		return Cond{}, lineErr(x, "unsupported comparison")
	}

	c.Value = v.Value

	return c, nil
}

// _relOps the numeric operator of each Lua comparison.
var _relOps = map[string]Op{
	">": OpGt, ">=": OpGte, "<": OpLt, "<=": OpLte,
}

// numeric tonumber(field) <op> n.
func (p *parser) numeric(x *ast.RelationalOpExpr) (Cond, bool) {
	op, ok := _relOps[x.Operator]
	call, isCall := x.Lhs.(*ast.FuncCallExpr)
	n, isNum := x.Rhs.(*ast.NumberExpr)
	if !ok || !isCall || !isNum || !_number.MatchString(n.Value) {
		return Cond{}, false
	}

	fn, ok := call.Func.(*ast.IdentExpr)
	if !ok || fn.Value != "tonumber" || len(call.Args) != 1 {
		return Cond{}, false
	}

	c, ok := p.field(call.Args[0])
	c.Op, c.Value = op, n.Value

	return c, ok
}

// affix field:sub(1, n) == "v" (prefix) or field:sub(-n) == "v"
// (suffix), n the length of v.
func (p *parser) affix(e ast.Expr, v string) (Cond, error) {
	call, ok := e.(*ast.FuncCallExpr)
	if !ok || call.Method != "sub" {
		return Cond{}, lineErr(e, "unsupported comparison")
	}

	c, ok := p.field(call.Receiver)
	if !ok || v == "" {
		return Cond{}, lineErr(e, "unsupported comparison")
	}

	n := strconv.Itoa(len(v))
	switch args := call.Args; {
	case len(args) == 2 && isNumber(args[0], "1") && isNumber(args[1], n):
		c.Op = OpPrefix
	case len(args) == 1 && isNegative(args[0], n):
		c.Op = OpSuffix
	default:
		return Cond{}, lineErr(e, "sub takes the length of the value: "+
			"field:sub(1, "+n+") or field:sub(-"+n+")")
	}

	c.Value = v

	return c, nil
}

// contains field:find("v", 1, true).
func (p *parser) contains(call *ast.FuncCallExpr) (Cond, error) {
	c, ok := p.field(call.Receiver)
	args := call.Args
	if !ok || call.Method != "find" || len(args) != 3 {
		return Cond{}, lineErr(call, "unsupported condition")
	}

	v, ok := args[0].(*ast.StringExpr)
	_, plain := args[2].(*ast.TrueExpr)
	if !ok || v.Value == "" || !isNumber(args[1], "1") || !plain {
		return Cond{}, lineErr(call,
			`find takes a plain search: field:find("v", 1, true)`)
	}

	c.Op, c.Value = OpContains, v.Value

	return c, nil
}

// field the source e reads: ctx:path(), ctx:header("k")...
func (p *parser) field(e ast.Expr) (Cond, bool) {
	call, ok := e.(*ast.FuncCallExpr)
	if !ok || call.Receiver == nil {
		return Cond{}, false
	}

	recv, ok := call.Receiver.(*ast.IdentExpr)
	if !ok || recv.Value != p.ctx {
		return Cond{}, false
	}

	for src, method := range _sourceMethods {
		if method == call.Method {
			return fieldOf(src, call.Args)
		}
	}

	return Cond{}, false
}

func fieldOf(src Source, args []ast.Expr) (Cond, bool) {
	if !src.Keyed() {
		return Cond{Source: src}, len(args) == 0
	}

	if len(args) != 1 {
		return Cond{}, false
	}

	key, ok := args[0].(*ast.StringExpr)
	if !ok || key.Value == "" {
		return Cond{}, false
	}

	return Cond{Source: src, Key: key.Value}, true
}

func isNumber(e ast.Expr, want string) bool {
	n, ok := e.(*ast.NumberExpr)

	return ok && n.Value == want
}

func isNegative(e ast.Expr, want string) bool {
	m, ok := e.(*ast.UnaryMinusOpExpr)

	return ok && isNumber(m.Expr, want)
}
