// Copyright 2026 Sentinéz Labs.
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
	"strings"
	"testing"

	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func strCond(src rulepb.FieldSource, op rulepb.Operator, v string,
) *rulepb.Rule {
	return strRule(src, "", op, structpb.NewStringValue(v))
}

// orOf a group of its rules, one of which matches.
func orOf(rules ...*rulepb.Rule) []*rulepb.AndCondition {
	ors := make([]*rulepb.AndCondition, 0, len(rules))
	for _, r := range rules {
		ors = append(ors, &rulepb.AndCondition{Rules: []*rulepb.Rule{r}})
	}

	return ors
}

// bothGroups (method one of two) and (path one of two): the second group
// is an and condition of the first.
func bothGroups(method, prefix string) *rulepb.Expression {
	return &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{
		OrCondition: orOf(
			strCond(_srcMethod, _eq, method),
			strCond(_srcMethod, _eq, "TRACE"),
		),
		AndCondition: []*rulepb.AndCondition{{
			OrCondition: orOf(
				strCond(_path, rulepb.Operator_OPERATOR_PREFIX, prefix),
				strCond(_path, _eq, "/never"),
			),
		}},
	}}}
}

// TestAndConditionEvaluates the request matches when each of the or
// groups does: PUT /api/test.
func TestAndConditionEvaluates(t *testing.T) {
	tests := []struct {
		name   string
		method string
		prefix string
		want   bool
	}{
		{"both", "PUT", "/api", true},
		{"first only", "PUT", "/other", false},
		{"second only", "GET", "/api", false},
		{"none", "GET", "/other", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eval := buildEval("r", bothGroups(tt.method, tt.prefix))
			assert.Equal(t, tt.want, eval(newTestContext(), nil))
		})
	}
}

func TestAndConditionWithRules(t *testing.T) {
	// the rules, the or group and the and group all need to match
	expr := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{
		Rules:       []*rulepb.Rule{strCond(_path, _eq, "/api/test")},
		OrCondition: orOf(strCond(_srcMethod, _eq, "PUT")),
		AndCondition: []*rulepb.AndCondition{
			{Rules: []*rulepb.Rule{strCond(_header, _eq, "x")}},
		},
	}}}

	assert.False(t, buildEval("r", expr)(newTestContext(), nil))

	expr.OrCondition[0].AndCondition[0].Rules[0] = strCond(
		_path, rulepb.Operator_OPERATOR_CONTAINS, "test")
	assert.True(t, buildEval("r", expr)(newTestContext(), nil))
}

func TestAndConditionString(t *testing.T) {
	want := `ctx:method() == "PUT" or ctx:method() == "TRACE"`
	got := ExprString(bothGroups("PUT", "/api"))

	assert.Equal(t, `(`+want+`) and (ctx:path():sub(1, 4) == "/api" or `+
		`ctx:path() == "/never")`, got)
}

func TestAndConditionText(t *testing.T) {
	invalid := strCond(_path, rulepb.Operator_OPERATOR_MATCHES, "(")
	expr := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{
		Rules: []*rulepb.Rule{invalid},
		AndCondition: []*rulepb.AndCondition{
			{OrCondition: orOf(invalid, invalid)},
		},
	}}}

	assert.Equal(t,
		`path matches "(" and (path matches "(" or path matches "(")`,
		exprText(expr))
}

func TestValidateAndCondition(t *testing.T) {
	require.NoError(t, ValidateExpression(bothGroups("PUT", "/api")))

	empty := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{
		AndCondition: []*rulepb.AndCondition{{}},
	}}}
	err := ValidateExpression(empty)
	require.ErrorIs(t, err, ErrInvalidCondition)
	assert.True(t, strings.HasPrefix(err.Error(),
		"orCondition[0].andCondition[0]: "), err.Error())

	bad := bothGroups("PUT", "/api")
	bad.OrCondition[0].AndCondition[0].OrCondition[0].Rules[0].
		Condition.Operator = rulepb.Operator_OPERATOR_GT
	err = ValidateExpression(bad)
	require.ErrorIs(t, err, ErrInvalidCondition)
	assert.Contains(t, err.Error(), "orCondition[0].andCondition[0]")
}

func TestValidateAndConditionLimits(t *testing.T) {
	rule := strCond(_path, _eq, "/a")

	deep := &rulepb.AndCondition{Rules: []*rulepb.Rule{rule}}
	for range _maxDepth {
		deep = &rulepb.AndCondition{AndCondition: []*rulepb.AndCondition{deep}}
	}

	err := ValidateExpression(&rulepb.Expression{
		OrCondition: []*rulepb.AndCondition{deep},
	})
	require.ErrorIs(t, err, ErrInvalidCondition)
	assert.Contains(t, err.Error(), "nested deeper")

	wide := &rulepb.AndCondition{}
	for range _maxConditions + 1 {
		wide.AndCondition = append(wide.AndCondition,
			&rulepb.AndCondition{Rules: []*rulepb.Rule{rule}})
	}

	err = ValidateExpression(&rulepb.Expression{
		OrCondition: []*rulepb.AndCondition{wide},
	})
	require.ErrorIs(t, err, ErrInvalidCondition)
	assert.Contains(t, err.Error(), "conditions, at most")
}
