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
	"maps"
	"net/netip"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/sentinez/scripting/internal/api"
)

// _number the values of the numeric conditions: a decimal literal Lua
// reads the same.
var _number = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// _ident a Lua name, a table key written without brackets.
var _ident = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// MatchSource the source of the secrule script of r: ParseMatch reads r
// back from it.
func MatchSource(r Rule) (string, error) {
	ret, err := actionLua(r.Action)
	if err != nil {
		return "", err
	}

	return emit(MatchFunc, r.Match, ret)
}

// KeySource the source of the rate limiter script counting, under the
// client IP, the requests match matches: ParseKey reads match back from
// it.
func KeySource(match Node) (string, error) {
	return emit(KeyFunc, match, "true")
}

// CondSource the Lua condition match matches, the if of the scripts
// MatchSource and KeySource write:
//
//	ctx:path() == "/login" and (ctx:header("X-A") ~= "" or ctx:ip() == "1")
//
// "and" binds tighter than "or" in Lua, as in a Node: an or group in an
// and group is put in parentheses. A condition matching every request is
// "true".
func CondSource(match Node) (string, error) {
	if err := checkNode(match); err != nil {
		return "", err
	}

	return nodeLua(match)
}

// checkNode that n has a branch and is not larger than MaxConditions.
func checkNode(n Node) error {
	if n.IsOr() && len(n.Nodes) == 0 {
		// the edge matches no request: no script says that
		return errNoBranch
	}

	if n.count() > MaxConditions {
		return errTooMany
	}

	return nil
}

func emit(fn string, match Node, ret string) (string, error) {
	if err := checkNode(match); err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("---@param ctx sz.Context\nfunction ")
	b.WriteString(fn)
	b.WriteString("(ctx)\n")

	for _, branch := range match.Branches() {
		if err := emitBranch(&b, branch, ret); err != nil {
			return "", err
		}

		if branch.matchesAll() {
			// the next branches never run
			break
		}
	}

	b.WriteString("end\n")

	return b.String(), nil
}

// emitBranch the if of branch, or the return of a branch matching every
// request.
func emitBranch(b *strings.Builder, branch Node, ret string) error {
	if branch.matchesAll() {
		b.WriteString("  return ")
		b.WriteString(ret)
		b.WriteString("\n")

		return nil
	}

	cond, err := nodeLua(branch)
	if err != nil {
		return err
	}

	b.WriteString("  if ")
	b.WriteString(cond)
	b.WriteString(" then\n    return ")
	b.WriteString(ret)
	b.WriteString("\n  end\n")

	return nil
}

// nodeLua the Lua expression of n.
func nodeLua(n Node) (string, error) {
	if n.Cond != nil {
		return condLua(*n.Cond)
	}

	if n.IsOr() && len(n.Nodes) == 0 {
		return "", errNoBranch
	}

	parts := make([]string, 0, len(n.Nodes))
	for _, c := range n.Nodes {
		if c.matchesAll() {
			if n.And {
				continue
			}

			// an or group with a branch matching every request does
			return "true", nil
		}

		s, err := nodeLua(c)
		if err != nil {
			return "", err
		}

		parts = append(parts, groupLua(n, c, s))
	}

	return joinLua(n.And, parts), nil
}

// groupLua the expression s of child, in parentheses in an and group when
// child is an or group: "and" binds tighter.
func groupLua(parent, child Node, s string) string {
	if parent.And && child.IsOr() {
		return "(" + s + ")"
	}

	return s
}

// joinLua the parts joined by "and", or by "or", "true" when there are none.
func joinLua(and bool, parts []string) string {
	switch {
	case len(parts) == 0:
		return "true"
	case and:
		return strings.Join(parts, " and ")
	}

	return strings.Join(parts, " or ")
}

// condLua the Lua expression of c.
func condLua(c Cond) (string, error) {
	switch {
	case c.Source == SourceTLS:
		return tlsLua(c)
	case c.Op == OpWithin || c.Op == OpNotWithin:
		return withinLua(c)
	case c.Op == OpExists || c.Op == OpAbsent:
		return existsLua(c)
	}

	field, err := fieldLua(c)
	if err != nil {
		return "", err
	}

	if op, ok := _numericOps[c.Op]; ok {
		if !_number.MatchString(c.Value) {
			return "", fmt.Errorf("%w: %q is no decimal number",
				ErrNotDeclarative, c.Value)
		}

		return "tonumber(" + field + ") " + op + " " + c.Value, nil
	}

	return stringCondLua(field, c)
}

func stringCondLua(field string, c Cond) (string, error) {
	v := quote(c.Value)
	n := strconv.Itoa(len(c.Value))

	if c.Value == "" && slices.Contains(
		[]Op{OpContains, OpPrefix, OpSuffix, OpMatches}, c.Op) {
		return "", fmt.Errorf("%w: empty value", ErrNotDeclarative)
	}

	if c.Source == SourceIP && (c.Op != OpEq && c.Op != OpNe ||
		strings.Contains(c.Value, "/")) {
		// ctx:ip() is an address: ip_in reads a prefix
		return "", fmt.Errorf("%w: ip %d %q", ErrNotDeclarative, c.Op,
			c.Value)
	}

	switch c.Op {
	case OpEq:
		return field + " == " + v, nil
	case OpNe:
		return field + " ~= " + v, nil
	case OpContains:
		return field + ":find(" + v + ", 1, true)", nil
	case OpPrefix:
		return field + ":sub(1, " + n + ") == " + v, nil
	case OpSuffix:
		return field + ":sub(-" + n + ") == " + v, nil
	case OpMatches:
		return matchesLua(field, c.Value)
	}

	return "", fmt.Errorf("%w: operator %d", ErrNotDeclarative, c.Op)
}

func matchesLua(field, pattern string) (string, error) {
	if _, err := regexp.Compile(pattern); err != nil {
		return "", fmt.Errorf("%w: %w", ErrNotDeclarative, err)
	}

	return "re.match(" + field + ", " + quote(pattern) + ")", nil
}

// tlsLua ctx:tls(), or not ctx:tls(), as c compares it to "true".
func tlsLua(c Cond) (string, error) {
	if c.Key != "" || c.Value != "true" && c.Value != "false" ||
		c.Op != OpEq && c.Op != OpNe {
		return "", fmt.Errorf("%w: tls %d %q", ErrNotDeclarative, c.Op,
			c.Value)
	}

	isTLS := (c.Op == OpEq) == (c.Value == "true")

	return not(!isTLS) + "ctx:tls()", nil
}

// withinLua ctx:ip_in("prefix"), or its negation.
func withinLua(c Cond) (string, error) {
	if _, err := netip.ParsePrefix(c.Value); err != nil ||
		c.Source != SourceIP || c.Key != "" {
		return "", fmt.Errorf("%w: ip in %q", ErrNotDeclarative, c.Value)
	}

	return not(c.Op == OpNotWithin) + "ctx:ip_in(" + quote(c.Value) + ")",
		nil
}

// existsLua ctx:has_header("key"), or its negation.
func existsLua(c Cond) (string, error) {
	if !c.Source.Keyed() || c.Key == "" || c.Value != "" {
		return "", fmt.Errorf("%w: presence of %q", ErrNotDeclarative,
			c.Key)
	}

	method := _presenceMethods[c.Source]

	return not(c.Op == OpAbsent) + "ctx:" + method + "(" + quote(c.Key) +
		")", nil
}

func not(neg bool) string {
	if neg {
		return "not "
	}

	return ""
}

// fieldLua the ctx call reading the source of c.
func fieldLua(c Cond) (string, error) {
	method, ok := _sourceMethods[c.Source]
	if !ok {
		return "", fmt.Errorf("%w: source %d", ErrNotDeclarative, c.Source)
	}

	if c.Source.Keyed() != (c.Key != "") {
		return "", fmt.Errorf("%w: key %q of %s", ErrNotDeclarative,
			c.Key, method)
	}

	if c.Key != "" {
		return "ctx:" + method + "(" + quote(c.Key) + ")", nil
	}

	return "ctx:" + method + "()", nil
}

// actionLua the value a secrule script returns for a: its constant
// (SZ_BLOCK), or a table with its params.
func actionLua(a Action) (string, error) {
	if !IsAction(a.Name) {
		return "", fmt.Errorf("%w: action %q", ErrNotDeclarative, a.Name)
	}

	if len(a.Params) == 0 {
		return api.ActionConst(a.Name), nil
	}

	params, err := tableLua(a.Params, 0)
	if err != nil {
		return "", err
	}

	return "{type = " + api.ActionConst(a.Name) + ", params = " + params +
		"}", nil
}

// tableLua the table constructor of m, its keys sorted.
func tableLua(m map[string]any, depth int) (string, error) {
	if depth >= _maxDepth || len(m) > _maxItems {
		return "", errTooLarge
	}

	fields := make([]string, 0, len(m))
	for _, k := range slices.Sorted(maps.Keys(m)) {
		v, err := valueLua(m[k], depth)
		if err != nil {
			return "", fmt.Errorf("%s: %w", k, err)
		}

		fields = append(fields, keyLua(k)+" = "+v)
	}

	return "{" + strings.Join(fields, ", ") + "}", nil
}

func valueLua(v any, depth int) (string, error) {
	switch x := v.(type) {
	case string:
		return quote(x), nil
	case bool:
		return strconv.FormatBool(x), nil
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64), nil
	case map[string]any:
		return tableLua(x, depth+1)
	}

	return "", fmt.Errorf("%w: value of type %T", ErrNotDeclarative, v)
}

func keyLua(k string) string {
	if _ident.MatchString(k) && !_luaKeywords[k] {
		return k
	}

	return "[" + quote(k) + "]"
}

// _luaKeywords the names a table key is not written bare with.
var _luaKeywords = map[string]bool{
	"and": true, "break": true, "do": true, "else": true, "elseif": true,
	"end": true, "false": true, "for": true, "function": true, "if": true,
	"in": true, "local": true, "nil": true, "not": true, "or": true,
	"repeat": true, "return": true, "then": true, "true": true,
	"until": true, "while": true, "goto": true,
}

// quote s as a Lua string literal: printable ASCII as is, the other bytes
// escaped in decimal, which every Lua reads.
func quote(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')

	for i := range len(s) {
		switch c := s[i]; {
		case c == '"' || c == '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c >= 0x20 && c < 0x7f:
			b.WriteByte(c)
		default:
			b.WriteByte('\\')
			fmt.Fprintf(&b, "%03d", c)
		}
	}

	b.WriteByte('"')

	return b.String()
}
