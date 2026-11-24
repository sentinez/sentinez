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

	return emit(MatchFunc, r.Any, ret)
}

// KeySource the source of the rate limiter script counting, under the
// client IP, the requests one of the branches of anyOf matches: ParseKey
// reads anyOf back from it.
func KeySource(anyOf [][]Cond) (string, error) {
	return emit(KeyFunc, anyOf, "true")
}

func emit(fn string, anyOf [][]Cond, ret string) (string, error) {
	if len(anyOf) == 0 {
		return "", errNoBranch
	}

	var b strings.Builder
	b.WriteString("---@param ctx sz.Context\nfunction ")
	b.WriteString(fn)
	b.WriteString("(ctx)\n")

	for _, branch := range anyOf {
		if len(branch) == 0 {
			// matches every request: the next branches never run
			b.WriteString("  return ")
			b.WriteString(ret)
			b.WriteString("\n")

			break
		}

		if err := emitBranch(&b, branch, ret); err != nil {
			return "", err
		}
	}

	b.WriteString("end\n")

	return b.String(), nil
}

func emitBranch(b *strings.Builder, branch []Cond, ret string) error {
	conds := make([]string, 0, len(branch))
	for _, c := range branch {
		s, err := condLua(c)
		if err != nil {
			return err
		}

		conds = append(conds, s)
	}

	b.WriteString("  if ")
	b.WriteString(strings.Join(conds, " and "))
	b.WriteString(" then\n    return ")
	b.WriteString(ret)
	b.WriteString("\n  end\n")

	return nil
}

// condLua the Lua expression of c.
func condLua(c Cond) (string, error) {
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
		[]Op{OpContains, OpPrefix, OpSuffix}, c.Op) {
		return "", fmt.Errorf("%w: empty value", ErrNotDeclarative)
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
	}

	return "", fmt.Errorf("%w: operator %d", ErrNotDeclarative, c.Op)
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
