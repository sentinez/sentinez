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

package securitysvc

import (
	"testing"

	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// pathExpr an expression with one condition on the path: PREFIX is in the
// schema, IN is not.
func pathExpr(op rulepb.Operator) *rulepb.Expression {
	return &rulepb.Expression{
		OrCondition: []*rulepb.AndCondition{{
			Rules: []*rulepb.Rule{{
				Condition: &rulepb.Condition{
					Source:   rulepb.FieldSource_FIELD_SOURCE_PATH,
					Operator: op,
					Value:    structpb.NewStringValue("/api"),
				},
			}},
		}},
	}
}

// invalidExpr an expression whose condition is outside of the schema.
func invalidExpr() *rulepb.Expression {
	return pathExpr(rulepb.Operator_OPERATOR_IN)
}

func TestValidateSecRule_Valid(t *testing.T) {
	tests := []struct {
		name string
		give *securitypb.SecRule
	}{
		{"no expr", &securitypb.SecRule{}},
		{"valid expr", &securitypb.SecRule{
			Expr: pathExpr(rulepb.Operator_OPERATOR_PREFIX),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, validateSecRule(tt.give))
		})
	}
}

func TestValidateSecRule_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		give    *securitypb.SecRule
		wantErr string
	}{
		{"nil", nil, "secrule is required"},
		{"invalid expr", &securitypb.SecRule{Expr: invalidExpr()},
			`expr: orCondition[0].rules[0]: invalid condition: ` +
				`operator "in" does not apply to path`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSecRule(tt.give)
			require.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestGetConditionSchema(t *testing.T) {
	resp, err := (&SecurityService{}).GetConditionSchema(t.Context(),
		&securitypb.GetConditionSchemaRequest{})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.GetSchema().GetSources())
}
