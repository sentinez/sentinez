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

package szrule

import (
	"slices"
	"strconv"
	"strings"

	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	_sourcePrefix   = "FIELD_SOURCE_"
	_operatorPrefix = "OPERATOR_"
	_matchAll       = "true"
)

// ExprString renders the expression in a human readable form, e.g.
//
//	(path eq "/login" and header["x-api-key"] ne "") or ip in ["10.0.0.1"]
//
// "and" binds tighter than "or", as in the evaluator. An empty
// expression, or an empty conjunction, matches every request and is
// rendered as "true".
func ExprString(expr *rulepb.Expression) string {
	var b strings.Builder

	writeOr(&b, expr.GetOrCondition())

	return b.String()
}

// writeOr writes the disjunction, grouping each of its conjunctions
// when there is more than one.
func writeOr(b *strings.Builder, or []*rulepb.AndCondition) {
	if len(or) == 0 {
		b.WriteString(_matchAll)
		return
	}

	for i, and := range or {
		if i > 0 {
			b.WriteString(" or ")
		}

		writeAnd(b, and, len(or) > 1)
	}
}

// writeAnd writes the rules of the conjunction followed by its nested
// disjunction. With group set, a conjunction of several terms is
// wrapped in parentheses.
func writeAnd(b *strings.Builder, and *rulepb.AndCondition, group bool) {
	rules, nested := and.GetRules(), and.GetOrCondition()

	terms := len(rules)
	if len(nested) > 0 {
		terms++
	}

	if terms == 0 {
		b.WriteString(_matchAll)
		return
	}

	if group && terms > 1 {
		b.WriteByte('(')
		defer b.WriteByte(')')
	}

	for i, r := range rules {
		if i > 0 {
			b.WriteString(" and ")
		}

		writeCondition(b, r.GetCondition())
	}

	if len(nested) == 0 {
		return
	}

	if len(rules) > 0 {
		b.WriteString(" and ")
	}

	writeNested(b, nested, len(rules) > 0)
}

// writeNested writes a nested disjunction. It is wrapped in
// parentheses when it follows other terms, since "and" binds tighter.
func writeNested(
	b *strings.Builder, nested []*rulepb.AndCondition, inAnd bool,
) {
	if !inAnd || len(nested) == 1 {
		writeOr(b, nested)
		return
	}

	b.WriteByte('(')
	writeOr(b, nested)
	b.WriteByte(')')
}

// writeCondition writes `source[key] operator value`. Without a key, or
// with the key named after the source, the condition is on the source
// itself, matching the evaluator.
func writeCondition(b *strings.Builder, c *rulepb.Condition) {
	source := enumName(c.GetSource().String(), _sourcePrefix)
	b.WriteString(source)

	if key := c.GetKey(); key != "" && key != source {
		b.WriteByte('[')
		b.WriteString(strconv.Quote(key))
		b.WriteByte(']')
	}

	b.WriteByte(' ')
	b.WriteString(enumName(c.GetOperator().String(), _operatorPrefix))
	b.WriteByte(' ')
	writeValue(b, c.GetValue())
}

// writeValue writes the value as a literal: strings are quoted, lists
// in brackets, structs in braces with their keys sorted.
func writeValue(b *strings.Builder, v *structpb.Value) {
	switch k := v.GetKind().(type) {
	case *structpb.Value_StringValue:
		b.WriteString(strconv.Quote(k.StringValue))
	case *structpb.Value_NumberValue:
		b.WriteString(strconv.FormatFloat(k.NumberValue, 'g', -1, 64))
	case *structpb.Value_BoolValue:
		b.WriteString(strconv.FormatBool(k.BoolValue))
	case *structpb.Value_ListValue:
		writeList(b, k.ListValue.GetValues())
	case *structpb.Value_StructValue:
		writeStruct(b, k.StructValue.GetFields())
	default:
		b.WriteString("null")
	}
}

func writeList(b *strings.Builder, values []*structpb.Value) {
	b.WriteByte('[')

	for i, v := range values {
		if i > 0 {
			b.WriteString(", ")
		}

		writeValue(b, v)
	}

	b.WriteByte(']')
}

func writeStruct(b *strings.Builder, fields map[string]*structpb.Value) {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}

	slices.Sort(keys)
	b.WriteByte('{')

	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(strconv.Quote(k))
		b.WriteString(": ")
		writeValue(b, fields[k])
	}

	b.WriteByte('}')
}

// enumName the lower case name of an enum value without its prefix,
// words separated by spaces: OPERATOR_NOT_IN is "not in".
func enumName(name, prefix string) string {
	name = strings.TrimPrefix(name, prefix)
	return strings.ReplaceAll(strings.ToLower(name), "_", " ")
}
