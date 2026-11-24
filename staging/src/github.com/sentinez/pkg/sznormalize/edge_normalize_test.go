// Copyright 2025 Duc-Hung Ho.
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

package sznormalize

import (
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	cdnpb "github.com/sentinez/sentinez/api/proto/sentinez/types/cdn/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestNormalizeEdgeSettingRuleIdFilled(t *testing.T) {
	// from proxy.yaml: no id
	rule := &secrulepb.SecRule{}
	normalizeSecRule(rule)

	assert.NotEmpty(t, rule.GetId())
}

func TestNormalizeEdgeSettingRuleIdKept(t *testing.T) {
	// from the database
	rule := &secrulepb.SecRule{Id: "srule_1"}
	normalizeSecRule(rule)

	assert.Equal(t, "srule_1", rule.GetId())
}

func TestNormalizeEdgeSettingLimiterNoIngress(t *testing.T) {
	lim := &edgepb.RateLimit{}
	normalizeLimiter(lim)

	assert.Nil(t, lim.GetIngress())
}

func TestNormalizeEdgeSettingLimiterDurations(t *testing.T) {
	// Durations are [timeWindow, timeout].
	tests := []struct {
		name string
		give [2]string
		want [2]string
	}{
		{
			name: "missing both uses default",
			want: [2]string{"5s", "5s"},
		},
		{
			name: "missing timeout only",
			give: [2]string{"1m", ""},
			want: [2]string{"1m", "5s"},
		},
		{
			name: "set values are kept",
			give: [2]string{"1m", "30s"},
			want: [2]string{"1m", "30s"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lim := &edgepb.RateLimit{
				TimeWindow: tt.give[0],
				Timeout:    tt.give[1],
			}
			normalizeLimiter(lim)

			got := [2]string{lim.GetTimeWindow(), lim.GetTimeout()}
			if got != tt.want {
				t.Errorf("durations = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeEdgeSettingCDN(t *testing.T) {
	cdn := &cdnpb.CDN{Rule: &cdnpb.Rule{
		Id: "c1", Expr: exprOf(&rulepb.Condition{Value: anyValue(t, 443)}),
	}}

	NormalizeEdgeSetting(&edgepb.Setting{
		Delivery: &edgepb.Delivery{Cdn: []*cdnpb.CDN{cdn}},
	})

	cond := conditionOf(cdn.GetRule().GetExpr())
	assert.Equal(t, "c1", cdn.GetRule().GetId())
	assert.NotEmpty(t, cond.GetId())
	assert.Equal(t, "443", cond.GetValue().GetStringValue())
}

func TestStringValue(t *testing.T) {
	tests := []struct {
		name string
		give any
		want any
	}{
		{name: "string", give: "/api", want: "/api"},
		{name: "number", give: 443, want: "443"},
		{name: "float", give: 1.5, want: "1.5"},
		{name: "bool", give: true, want: "true"},
		{
			name: "list",
			give: []any{"GET", 1},
			want: []any{"GET", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringValue(anyValue(t, tt.give))
			assert.Equal(t, tt.want, got.AsInterface())
		})
	}
}

func TestStringValueNil(t *testing.T) {
	assert.Nil(t, stringValue(nil))
}

func TestResourceDomain(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{"lower case", "shop.s6z.io.vn", "shop.s6z.io.vn"},
		{"upper case", "Shop.S6Z.io.vn", "shop.s6z.io.vn"},
		{"trailing dot", "shop.s6z.io.vn.", "shop.s6z.io.vn"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ResourceDomain(tt.give))
		})
	}
}

// A rule rebuilt from the database keeps its id and its conditions.
func TestNormalizeKeepsDatabaseRule(t *testing.T) {
	cond := &rulepb.Condition{
		Id: "cond_1", Value: structpb.NewStringValue("POST"),
	}
	runtime := &secrulepb.SecRule{Id: "r1", Expr: exprOf(cond)}
	setting := &edgepb.Setting{Security: &edgepb.Security{
		Secrules: []*edgepb.SecRule{{Ingress: runtime}},
		Limiters: []*edgepb.RateLimit{{Ingress: runtime}},
	}}

	NormalizeEdgeSetting(setting)

	sec := setting.GetSecurity()
	assert.Same(t, runtime, sec.GetSecrules()[0].GetIngress())
	assert.Equal(t, "r1", runtime.GetId())
	assert.Equal(t, "cond_1", cond.GetId())
	assert.Equal(t, "5s", sec.GetLimiters()[0].GetTimeWindow())
}

func exprOf(cond *rulepb.Condition) *rulepb.Expression {
	return &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{
		Rules: []*rulepb.Rule{{Condition: cond}},
	}}}
}

func conditionOf(expr *rulepb.Expression) *rulepb.Condition {
	return expr.GetOrCondition()[0].GetRules()[0].GetCondition()
}

func anyValue(t *testing.T, v any) *structpb.Value {
	t.Helper()

	val, err := structpb.NewValue(v)
	require.NoError(t, err)

	return val
}

// TestNormalizeAndConditions the conditions of the and groups get their id
// and string value like the ones of the or groups.
func TestNormalizeAndConditions(t *testing.T) {
	inAnd := &rulepb.Condition{Value: anyValue(t, 1)}
	inOr := &rulepb.Condition{Value: anyValue(t, 2)}
	rule := &secrulepb.SecRule{Expr: &rulepb.Expression{
		OrCondition: []*rulepb.AndCondition{{
			AndCondition: []*rulepb.AndCondition{{
				Rules: []*rulepb.Rule{{Condition: inAnd}},
				AndCondition: []*rulepb.AndCondition{{
					OrCondition: []*rulepb.AndCondition{
						{Rules: []*rulepb.Rule{{Condition: inOr}}},
					},
				}},
			}},
		}},
	}}

	normalizeSecRule(rule)

	assert.NotEmpty(t, inAnd.GetId())
	assert.NotEmpty(t, inOr.GetId())
	assert.Equal(t, "1", inAnd.GetValue().GetStringValue())
	assert.Equal(t, "2", inOr.GetValue().GetStringValue())
}
