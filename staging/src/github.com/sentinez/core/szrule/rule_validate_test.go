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
	"testing"

	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	_srcQuery  = rulepb.FieldSource_FIELD_SOURCE_QUERY
	_srcTls    = rulepb.FieldSource_FIELD_SOURCE_TLS
	_srcMethod = rulepb.FieldSource_FIELD_SOURCE_METHOD

	_opGt      = rulepb.Operator_OPERATOR_GT
	_opIn      = rulepb.Operator_OPERATOR_IN
	_opMatches = rulepb.Operator_OPERATOR_MATCHES
	_opPrefix  = rulepb.Operator_OPERATOR_PREFIX
)

func cond(source rulepb.FieldSource, key string, op rulepb.Operator,
	value *structpb.Value) *rulepb.Condition {

	return &rulepb.Condition{
		Source: source, Key: key, Operator: op, Value: value,
	}
}

func str(s string) *structpb.Value {
	return structpb.NewStringValue(s)
}

func strs(values ...string) *structpb.Value {
	list := make([]*structpb.Value, 0, len(values))
	for _, v := range values {
		list = append(list, str(v))
	}

	return structpb.NewListValue(&structpb.ListValue{Values: list})
}

func TestValidateConditionValid(t *testing.T) {
	tests := []struct {
		name string
		give *rulepb.Condition
	}{
		{"header value", cond(_header, "User-Agent", _eq, str("curl"))},
		{"header number", cond(_header, "Content-Length", _opGt,
			str("1024"))},
		{"header in", cond(_header, "X-Env", _opIn, strs("a", "b"))},
		{"header presence", cond(_header, "", _opIn,
			strs("Cookie", "X-Api-Key"))},
		{"header legacy presence key", cond(_header, "header", _eq,
			str("Cookie"))},
		{"query value", cond(_srcQuery, "page", _eq, str("1"))},
		{"path regex", cond(_path, "", _opMatches, str(`^/api/v\d+`))},
		{"ip prefix", cond(_ip, "", _eq, str("10.0.0.0/8"))},
		{"ip list", cond(_ip, "", _opIn, strs("10.0.0.1", "::1"))},
		{"tls", cond(_srcTls, "", _eq, str("true"))},
		{"method list", cond(_srcMethod, "", _opIn, strs("GET", "POST"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, ValidateCondition(tt.give))
		})
	}
}

type invalidConditionTest struct {
	name    string
	give    *rulepb.Condition
	wantErr string
}

func TestValidateConditionInvalid(t *testing.T) {
	tests := slices.Concat(invalidKeyTests(), invalidValueTests())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCondition(tt.give)
			require.ErrorIs(t, err, ErrInvalidCondition)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

// invalidKeyTests a source, key or operator outside of the schema.
func invalidKeyTests() []invalidConditionTest {
	return []invalidConditionTest{
		{"missing", nil, "missing"},
		{"unspecified source", &rulepb.Condition{}, "unsupported source"},
		{"key on path", cond(_path, "x", _eq, str("/")),
			`path takes no key, got "x"`},
		{"bad header name", cond(_header, "User Agent", _eq,
			str("x")), `key: "User Agent" is not a valid header name`},
		{"operator of another source", cond(_path, "", _opIn,
			strs("/")), `operator "in" does not apply to path`},
		{"number on presence", cond(_header, "", _opGt, str("1")),
			`operator "gt" does not apply to header presence`},
		{"number op on path", cond(_path, "", _opGt, str("1")),
			`operator "gt" does not apply to path`},
		{"tls in", cond(_srcTls, "", _opIn, strs("true")),
			`operator "in" does not apply to tls`},
	}
}

// invalidValueTests a value of the wrong type or kind.
func invalidValueTests() []invalidConditionTest {
	return []invalidConditionTest{
		{"list for a scalar", cond(_path, "", _opPrefix, strs("/")),
			"the value must be a string"},
		{"string for a list", cond(_ip, "", _opIn, str("10.0.0.1")),
			"the value must be a list of strings"},
		{"json number", cond(_header, "Content-Length", _opGt,
			structpb.NewNumberValue(1)), "the value must be a string"},
		{"no value", cond(_path, "", _eq, nil),
			"the value must be a string"},
		{"empty value", cond(_path, "", _eq, str("")),
			"the value must be 1 to 4096 bytes"},
		{"empty list", cond(_ip, "", _opIn, strs()),
			"the list must have 1 to 256 values, got 0"},
		{"not a number", cond(_header, "Content-Length", _opGt,
			str("NaN")), `"NaN" is not a number`},
		{"bad regex", cond(_path, "", _opMatches, str("(")),
			`"(" is not a regular expression`},
		{"bad ip", cond(_ip, "", _opIn, strs("10.0.0.1", "nope")),
			`value 1: "nope" is not an IP address or prefix`},
		{"bad method", cond(_srcMethod, "", _eq, str("get")),
			`"get" is not one of GET, HEAD`},
		{"bad tls", cond(_srcTls, "", _eq, str("yes")),
			`"yes" is not one of true, false`},
	}
}

func TestValidateExpression(t *testing.T) {
	valid := strRule(_path, "", _opPrefix, str("/api"))
	invalid := strRule(_path, "", _opIn, str("/api"))
	tests := []struct {
		name    string
		give    *rulepb.Expression
		wantErr string
	}{
		{"nested invalid", &rulepb.Expression{
			OrCondition: []*rulepb.AndCondition{
				{Rules: []*rulepb.Rule{valid}},
				{
					Rules: []*rulepb.Rule{valid},
					OrCondition: []*rulepb.AndCondition{
						{Rules: []*rulepb.Rule{valid, invalid}},
					},
				},
			},
		}, `orCondition[1].orCondition[0].rules[1]: invalid condition`},
		{"empty group", &rulepb.Expression{
			OrCondition: []*rulepb.AndCondition{{}},
		}, "orCondition[0]: invalid condition: empty group"},
		{"too deep", nested(valid, _maxDepth+1),
			"groups nested deeper than 8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExpression(tt.give)
			require.ErrorIs(t, err, ErrInvalidCondition)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestValidateExpressionValid(t *testing.T) {
	valid := strRule(_path, "", _opPrefix, str("/api"))

	tests := []struct {
		name string
		give *rulepb.Expression
	}{
		{"nil", nil},
		{"empty", &rulepb.Expression{}},
		{"deepest", nested(valid, _maxDepth)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, ValidateExpression(tt.give))
		})
	}
}

// nested an expression of depth groups, each nested in the previous one.
func nested(rule *rulepb.Rule, depth int) *rulepb.Expression {
	group := &rulepb.AndCondition{Rules: []*rulepb.Rule{rule}}
	for range depth - 1 {
		group = &rulepb.AndCondition{
			Rules:       []*rulepb.Rule{rule},
			OrCondition: []*rulepb.AndCondition{group},
		}
	}

	return &rulepb.Expression{OrCondition: []*rulepb.AndCondition{group}}
}

// TestSchemaAccepted every operator of Schema accepts a value of its kind:
// the console may offer all of them.
func TestSchemaAccepted(t *testing.T) {
	for _, src := range Schema().GetSources() {
		for _, v := range src.GetVariants() {
			key := ""
			if v.GetKeyMode() == rulepb.KeyMode_KEY_MODE_REQUIRED {
				key = "X-Key"
			}

			for _, op := range v.GetOperators() {
				c := cond(src.GetSource(), key, op.GetOperator(),
					sample(op))
				assert.NoError(t, ValidateCondition(c), "%s %s %s",
					src.GetSource(), v.GetName(), op.GetOperator())
			}
		}
	}
}

func sample(op *rulepb.OperatorSpec) *structpb.Value {
	s := map[rulepb.ValueKind]string{
		rulepb.ValueKind_VALUE_KIND_STRING: "a",
		rulepb.ValueKind_VALUE_KIND_NUMBER: "1.5",
		rulepb.ValueKind_VALUE_KIND_REGEX:  "^a+$",
		rulepb.ValueKind_VALUE_KIND_IP:     "192.168.0.0/16",
		rulepb.ValueKind_VALUE_KIND_NAME:   "X-Name",
	}[op.GetValueKind()]
	if enum := op.GetEnumValues(); len(enum) > 0 {
		s = enum[0]
	}

	if op.GetList() {
		return strs(s)
	}

	return str(s)
}

func TestSchemaCopy(t *testing.T) {
	first := Schema()
	first.GetSources()[0].GetVariants()[0].Operators = nil

	i := slices.IndexFunc(Schema().GetSources(),
		func(s *rulepb.SourceSpec) bool {
			return s.GetSource() == _srcMethod
		})
	require.GreaterOrEqual(t, i, 0)

	method := Schema().GetSources()[i]
	method.GetVariants()[0].GetOperators()[0].EnumValues[0] = "X"

	assert.NotEmpty(t, Schema().GetSources()[0].GetVariants()[0].
		GetOperators())
	assert.Equal(t, "GET", _methods[0])
}

// TestSchemaCoversEnums a source or an operator added to the protos has
// to be listed in the schema, or ruled out here: the matchers evaluate
// them, and the validator would reject every rule using them.
func TestSchemaCoversEnums(t *testing.T) {
	ops := make(map[rulepb.Operator]bool)

	for _, src := range Schema().GetSources() {
		for _, v := range src.GetVariants() {
			for _, op := range v.GetOperators() {
				ops[op.GetOperator()] = true
			}
		}
	}

	for n, name := range rulepb.FieldSource_name {
		if n == int32(rulepb.FieldSource_FIELD_SOURCE_UNSPECIFIED) {
			continue
		}

		_, ok := findSource(rulepb.FieldSource(n))
		assert.True(t, ok, "source %s is not in the schema", name)
	}

	for n, name := range rulepb.Operator_name {
		if n == int32(rulepb.Operator_OPERATOR_UNSPECIFIED) {
			continue
		}

		assert.True(t, ops[rulepb.Operator(n)],
			"operator %s is not in the schema", name)
	}
}
