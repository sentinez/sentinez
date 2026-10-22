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
	"testing"

	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	_path   = rulepb.FieldSource_FIELD_SOURCE_PATH
	_header = rulepb.FieldSource_FIELD_SOURCE_HEADER
	_ip     = rulepb.FieldSource_FIELD_SOURCE_IP
	_eq     = rulepb.Operator_OPERATOR_EQ
	_notIn  = rulepb.Operator_OPERATOR_NOT_IN
)

func strRule(
	src rulepb.FieldSource, key string, op rulepb.Operator, v *structpb.Value,
) *rulepb.Rule {
	return &rulepb.Rule{Condition: &rulepb.Condition{
		Source: src, Key: key, Operator: op, Value: v,
	}}
}

func and(
	or []*rulepb.AndCondition, rules ...*rulepb.Rule,
) *rulepb.AndCondition {
	return &rulepb.AndCondition{Rules: rules, OrCondition: or}
}

// nolint:funlen
func TestExprString(t *testing.T) {
	login := strRule(_path, "", _eq, structpb.NewStringValue("/login"))
	apiKey := strRule(
		_header, "x-api-key", _eq, structpb.NewStringValue("k"),
	)
	ips, err := structpb.NewList([]any{"10.0.0.1", "10.0.0.2"})
	require.NoError(t, err)
	notIps := strRule(_ip, "", _notIn, structpb.NewListValue(ips))

	tests := []struct {
		name string
		give *rulepb.Expression
		want string
	}{
		{name: "nil", want: "true"},
		{
			name: "empty conjunction",
			give: &rulepb.Expression{
				OrCondition: []*rulepb.AndCondition{and(nil)},
			},
			want: "true",
		},
		{
			name: "single conjunction",
			give: &rulepb.Expression{OrCondition: []*rulepb.AndCondition{
				and(nil, login, apiKey),
			}},
			want: `path eq "/login" and header["x-api-key"] eq "k"`,
		},
		{
			name: "disjunction",
			give: &rulepb.Expression{OrCondition: []*rulepb.AndCondition{
				and(nil, login, apiKey), and(nil, notIps),
			}},
			want: `(path eq "/login" and header["x-api-key"] eq "k")` +
				` or ip not in ["10.0.0.1", "10.0.0.2"]`,
		},
		{
			name: "nested disjunction",
			give: &rulepb.Expression{OrCondition: []*rulepb.AndCondition{
				and([]*rulepb.AndCondition{
					and(nil, apiKey), and(nil, notIps),
				}, login),
			}},
			want: `path eq "/login" and (header["x-api-key"] eq "k"` +
				` or ip not in ["10.0.0.1", "10.0.0.2"])`,
		},
		{
			name: "key named after source",
			give: &rulepb.Expression{OrCondition: []*rulepb.AndCondition{
				and(nil, strRule(
					_header, "header", _eq, structpb.NewNumberValue(1.5),
				)),
			}},
			want: `header eq 1.5`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExprString(tt.give))
		})
	}
}
