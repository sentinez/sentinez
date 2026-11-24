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
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuin/gopher-lua/ast"
)

// cond the condition e, one of the comparisons MatchSource writes.
func (p *parser) cond(e ast.Expr) (Cond, error) {
	switch x := e.(type) {
	case *ast.RelationalOpExpr:
		return p.relational(x)
	case *ast.FuncCallExpr:
		return p.call(x)
	}

	return Cond{}, lineErr(e, "unsupported condition")
}

// call a condition of a boolean call: field:find("v", 1, true),
// re.match(field, "v"), ctx:tls(), ctx:ip_in("cidr"), ctx:has_header("k")
// or ctx:has_query("k").
func (p *parser) call(x *ast.FuncCallExpr) (Cond, error) {
	if isRe(x.Func) {
		return p.matches(x)
	}

	recv, ok := x.Receiver.(*ast.IdentExpr)
	if !ok || recv.Value != p.ctx {
		return p.contains(x)
	}

	switch x.Method {
	case "tls":
		if len(x.Args) == 0 {
			return Cond{Source: SourceTLS, Op: OpEq, Value: "true"}, nil
		}
	case "ip_in":
		return ipIn(x)
	}

	for src, method := range _presenceMethods {
		if method == x.Method {
			return presence(x, src)
		}
	}

	return Cond{}, lineErr(x, "unsupported condition")
}

// isRe whether e is re.match.
func isRe(e ast.Expr) bool {
	get, ok := e.(*ast.AttrGetExpr)
	if !ok {
		return false
	}

	obj, isIdent := get.Object.(*ast.IdentExpr)
	key, isStr := get.Key.(*ast.StringExpr)

	return isIdent && isStr && obj.Value == "re" && key.Value == "match"
}

// matches re.match(field, "pattern").
func (p *parser) matches(x *ast.FuncCallExpr) (Cond, error) {
	if len(x.Args) != 2 {
		return Cond{}, lineErr(x, `re.match(field, "pattern") expected`)
	}

	c, ok := p.field(x.Args[0])
	v, isStr := x.Args[1].(*ast.StringExpr)
	if !ok || !isStr || v.Value == "" {
		return Cond{}, lineErr(x, `re.match(field, "pattern") expected`)
	}

	if _, err := regexp.Compile(v.Value); err != nil {
		return Cond{}, lineErr(x, err.Error())
	}

	c.Op, c.Value = OpMatches, v.Value

	return c, nil
}

// ipIn ctx:ip_in("cidr").
func ipIn(x *ast.FuncCallExpr) (Cond, error) {
	v, ok := stringArg(x)
	if !ok {
		return Cond{}, lineErr(x, `ctx:ip_in("cidr") expected`)
	}

	if _, err := netip.ParsePrefix(v); err != nil {
		return Cond{}, lineErr(x, err.Error())
	}

	return Cond{Source: SourceIP, Op: OpWithin, Value: v}, nil
}

// presence ctx:has_header("k") or ctx:has_query("k").
func presence(x *ast.FuncCallExpr, src Source) (Cond, error) {
	k, ok := stringArg(x)
	if !ok || k == "" {
		return Cond{}, lineErr(x, "ctx:"+x.Method+`("name") expected`)
	}

	return Cond{Source: src, Key: k, Op: OpExists}, nil
}

// stringArg the string literal x takes as its only argument.
func stringArg(x *ast.FuncCallExpr) (string, bool) {
	if len(x.Args) != 1 {
		return "", false
	}

	v, ok := x.Args[0].(*ast.StringExpr)
	if !ok {
		return "", false
	}

	return v.Value, true
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
	case ok && c.Source == SourceIP && strings.Contains(v.Value, "/"):
		return Cond{}, lineErr(x, `ctx:ip() is an address: `+
			`ctx:ip_in("cidr") reads a prefix`)
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

// _relOps the numeric operator of each Lua comparison: _numericOps
// inverted.
var _relOps = func() map[string]Op {
	m := make(map[string]Op, len(_numericOps))
	for op, rel := range _numericOps {
		m[rel] = op
	}

	return m
}()

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
		// ctx:tls() is a boolean, a condition by itself
		if method == call.Method && src != SourceTLS {
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
