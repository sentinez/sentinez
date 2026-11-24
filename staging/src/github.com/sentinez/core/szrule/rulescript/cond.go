// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rulescript

import (
	"fmt"
	"net/netip"

	"github.com/sentinez/scripting"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// _sources the script source of each field source.
var _sources = map[rulepb.FieldSource]scripting.Source{
	rulepb.FieldSource_FIELD_SOURCE_PATH:   scripting.SourcePath,
	rulepb.FieldSource_FIELD_SOURCE_METHOD: scripting.SourceMethod,
	rulepb.FieldSource_FIELD_SOURCE_HOST:   scripting.SourceHost,
	rulepb.FieldSource_FIELD_SOURCE_IP:     scripting.SourceIP,
	rulepb.FieldSource_FIELD_SOURCE_HEADER: scripting.SourceHeader,
	rulepb.FieldSource_FIELD_SOURCE_QUERY:  scripting.SourceQuery,
	rulepb.FieldSource_FIELD_SOURCE_BODY:   scripting.SourceBody,
	rulepb.FieldSource_FIELD_SOURCE_JA4:    scripting.SourceJA4,
	rulepb.FieldSource_FIELD_SOURCE_TLS:    scripting.SourceTLS,
}

// _ops the script operator of each operator of a single value. IN, NOT_IN
// and the lists become groups of conditions (listNode).
var _ops = map[rulepb.Operator]scripting.Op{
	rulepb.Operator_OPERATOR_EQ:       scripting.OpEq,
	rulepb.Operator_OPERATOR_NE:       scripting.OpNe,
	rulepb.Operator_OPERATOR_CONTAINS: scripting.OpContains,
	rulepb.Operator_OPERATOR_PREFIX:   scripting.OpPrefix,
	rulepb.Operator_OPERATOR_SUFFIX:   scripting.OpSuffix,
	rulepb.Operator_OPERATOR_GT:       scripting.OpGt,
	rulepb.Operator_OPERATOR_GTE:      scripting.OpGte,
	rulepb.Operator_OPERATOR_LT:       scripting.OpLt,
	rulepb.Operator_OPERATOR_LTE:      scripting.OpLte,
	rulepb.Operator_OPERATOR_MATCHES:  scripting.OpMatches,
}

// _presenceKeys the key, besides the empty one, of the header and query
// conditions checking which fields are present (szrule), not a value.
var _presenceKeys = map[scripting.Source]string{
	scripting.SourceHeader: "header",
	scripting.SourceQuery:  "query",
}

// nodeOf the condition of expr: the edge matches a request when one of
// its or conditions does, an empty expr matching every request.
func nodeOf(expr *rulepb.Expression) (scripting.Node, error) {
	return orNode(expr.GetOrCondition())
}

// orNode the node matching when one of ors does, every request when there
// is none.
func orNode(ors []*rulepb.AndCondition) (scripting.Node, error) {
	if len(ors) == 0 {
		return scripting.All(), nil
	}

	nodes := make([]scripting.Node, 0, len(ors))
	for _, and := range ors {
		n, err := andNode(and)
		if err != nil {
			return scripting.Node{}, err
		}

		nodes = append(nodes, n)
	}

	return scripting.Any(nodes...), nil
}

// andNode the node matching when the rules, the or group and every and
// group of and do.
func andNode(and *rulepb.AndCondition) (scripting.Node, error) {
	nodes := make([]scripting.Node, 0, len(and.GetRules())+1)

	for _, r := range and.GetRules() {
		n, err := condNode(r.GetCondition())
		if err != nil {
			return scripting.Node{}, err
		}

		nodes = append(nodes, n)
	}

	if len(and.GetOrCondition()) > 0 {
		n, err := orNode(and.GetOrCondition())
		if err != nil {
			return scripting.Node{}, err
		}

		nodes = append(nodes, n)
	}

	for _, sub := range and.GetAndCondition() {
		n, err := andNode(sub)
		if err != nil {
			return scripting.Node{}, err
		}

		nodes = append(nodes, n)
	}

	return scripting.All(nodes...), nil
}

// condNode the node of c: one condition, or the group of a list of values
// matched one of (EQ, IN) or none of (NE, NOT_IN).
func condNode(c *rulepb.Condition) (scripting.Node, error) {
	if c == nil {
		return scripting.Node{}, errNoCondition
	}

	field, err := fieldOf(c)
	if err != nil {
		return scripting.Node{}, err
	}

	if isPresence(field.Source, c.GetKey()) {
		return presenceNode(field.Source, c)
	}

	if list := c.GetValue().GetListValue(); list != nil {
		return listNode(field, c.GetOperator(), list.GetValues())
	}

	return single(field, c)
}

// fieldOf the source and key c reads.
func fieldOf(c *rulepb.Condition) (scripting.Cond, error) {
	src, ok := _sources[c.GetSource()]
	if !ok {
		return scripting.Cond{}, unsupported("source %s", c.GetSource())
	}

	field := scripting.Cond{Source: src}
	if src.Keyed() && !isPresence(src, c.GetKey()) {
		field.Key = c.GetKey()
	}

	return field, nil
}

// isPresence whether the condition of src keyed by key checks which
// fields the request has.
func isPresence(src scripting.Source, key string) bool {
	return src.Keyed() && (key == "" || key == _presenceKeys[src])
}

func single(field scripting.Cond, c *rulepb.Condition,
) (scripting.Node, error) {
	v, ok := c.GetValue().GetKind().(*structpb.Value_StringValue)
	if !ok {
		return scripting.Node{}, unsupported("value of %s: not a string",
			c.GetSource())
	}

	op, ok := _ops[c.GetOperator()]
	if !ok {
		return scripting.Node{}, unsupported("operator %s", c.GetOperator())
	}

	field.Op, field.Value = op, v.StringValue
	if field.Source == scripting.SourceIP {
		field.Op, ok = ipOp(op, v.StringValue)
		if !ok {
			return scripting.Node{}, unsupported("ip %s", c.GetOperator())
		}
	}

	return scripting.Leaf(field), nil
}

// ipOp the op of an ip condition: szrule reads a value that parses as a
// CIDR prefix as one (matchIP), and compares with EQ and NE only.
func ipOp(op scripting.Op, value string) (scripting.Op, bool) {
	_, err := netip.ParsePrefix(value)
	cidr := err == nil
	switch {
	case op == scripting.OpEq && cidr:
		return scripting.OpWithin, true
	case op == scripting.OpNe && cidr:
		return scripting.OpNotWithin, true
	}

	return op, op == scripting.OpEq || op == scripting.OpNe
}

// presenceNode the node of a presence condition of src: the request has
// the named field, or each of a list of them (EQ, IN), or none (NE,
// NOT_IN).
func presenceNode(src scripting.Source, c *rulepb.Condition,
) (scripting.Node, error) {
	op := scripting.OpExists
	switch c.GetOperator() {
	case _ne, _notIn:
		op = scripting.OpAbsent
	case _eq, _in:
	default:
		return scripting.Node{}, unsupported("presence %s", c.GetOperator())
	}

	values := []*structpb.Value{c.GetValue()}
	if list := c.GetValue().GetListValue(); list != nil {
		values = list.GetValues()
	}

	nodes := make([]scripting.Node, 0, len(values))
	for _, v := range values {
		name := v.GetStringValue()
		if name == "" {
			return scripting.Node{}, unsupported("presence of %v", v)
		}

		nodes = append(nodes, scripting.Leaf(
			scripting.Cond{Source: src, Key: name, Op: op}))
	}

	return scripting.All(nodes...), nil
}

func unsupported(format string, args ...any) error {
	return fmt.Errorf("%w: "+format,
		append([]any{scripting.ErrNotDeclarative}, args...)...)
}

// _sourcePbs and _opPbs the field source and operator of each script
// one: _sources and _ops inverted.
var (
	_sourcePbs = invert(_sources)
	_opPbs     = invert(_ops)
)

// condPb the condition of c.
func condPb(c scripting.Cond) *rulepb.Condition {
	cond := &rulepb.Condition{
		Source:   _sourcePbs[c.Source],
		Key:      c.Key,
		Operator: _opPbs[c.Op],
		Value:    structpb.NewStringValue(c.Value),
	}

	switch c.Op {
	case scripting.OpWithin, scripting.OpExists:
		cond.Operator = _eq
	case scripting.OpNotWithin, scripting.OpAbsent:
		cond.Operator = _ne
	}

	if c.Op == scripting.OpExists || c.Op == scripting.OpAbsent {
		// the name of the field is the value of a condition without key
		cond.Key, cond.Value = "", structpb.NewStringValue(c.Key)
	}

	return cond
}

func invert[K, V comparable](m map[K]V) map[V]K {
	out := make(map[V]K, len(m))
	for k, v := range m {
		out[v] = k
	}

	return out
}
