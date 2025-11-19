// Copyright 2025 Duc-Hung Ho.
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

package corerule

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	corehttp "github.com/sentinez/core/http"
	corehttpreq "github.com/sentinez/core/http/request"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// nolint
func newBaseContext() *edgepb.RequestContext {
	return &edgepb.RequestContext{
		Body: []byte(`{"username":"hung","password":"123456"}`),
		Header: map[string]string{
			"Content-Type":  "application/json",
			"User-Agent":    "curl/8.0.1",
			"Accept":        "*/*",
			"Authorization": "Bearer abc.def.ghi",
		},
		Host:   "api.example.com",
		Ip:     "203.0.113.42",
		Ja4:    "ja4:abcd1234efgh5678ijkl9012mnop3456",
		Method: "POST",
		Path:   "/v1/login",
		Queries: map[string]*edgepb.RequestQuery{
			"lang":  {Value: []string{"vi"}},
			"lang2": {Value: []string{"vi"}},
		},
		Tls:           true,
		Protocol:      "HTTP/1.1",
		RemoteAddress: "203.0.113.42:52341",
		StatusCode:    200,
		Uri:           "https://api.example.com/v1/login?redirect=/home&lang=en",
	}
}

//nolint:lll
func newContext() corehttp.RequestContext {
	reqCtx := newBaseContext()
	return corehttpreq.NewRequestContext(context.Background(), reqCtx)
}

func TestRulePath(t *testing.T) {
	rule := NewIngress()

	req := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_PATH,
			Operator: ruleenginepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("/v1/login"),
			Key:      "path",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %s", string(val))

	ok := rule.Eval(newContext(), req)
	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

func TestRuleQuery(t *testing.T) {
	rule := NewIngress()

	req := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_QUERY,
			Operator: ruleenginepb.Operator_OPERATOR_IN,
			Value: structpb.NewListValue(&structpb.ListValue{
				Values: []*structpb.Value{
					structpb.NewStringValue("lang"),
					structpb.NewStringValue("lang2"),
				},
			}),
			Key: "query",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := rule.Eval(newContext(), req)

	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

func TestRuleClientIP(t *testing.T) {
	rule := NewIngress()

	req := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_IP,
			Operator: ruleenginepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("203.0.113.42"),
			Key:      "ip",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := rule.Eval(newContext(), req)

	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

func TestRuleClientIPRange(t *testing.T) {
	rule := NewIngress()

	req := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_IP,
			Operator: ruleenginepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("203.0.113.0/24"),
			Key:      "ip",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := rule.Eval(newContext(), req)

	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

func TestRuleClientIPRangeNotEQ(t *testing.T) {
	rule := NewIngress()

	req := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_IP,
			Operator: ruleenginepb.Operator_OPERATOR_NE,
			Value:    structpb.NewStringValue("203.1.113.0/24"),
			Key:      "ip",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := rule.Eval(newContext(), req)

	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

// nolint
func TestChain(t *testing.T) {
	rulePath := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_PATH,
			Operator: ruleenginepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("/v1/login"),
			Key:      "path",
		},
	}

	ruleQuery := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_QUERY,
			Operator: ruleenginepb.Operator_OPERATOR_IN,
			Value: structpb.NewListValue(&structpb.ListValue{
				Values: []*structpb.Value{
					structpb.NewStringValue("lang"),
					structpb.NewStringValue("lang2"),
				},
			}),
			Key: "query",
		},
	}

	ruleClientIP := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_IP,
			Operator: ruleenginepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("203.0.113.42"),
			Key:      "ip",
		},
	}

	ruleChain := &ruleenginepb.Expr{
		Enabled: true,
		Rules:   []*ruleenginepb.Rule{rulePath, ruleQuery, ruleClientIP, ruleClientIP},
		Logics:  []ruleenginepb.Logic{ruleenginepb.Logic_LOGIC_AND, ruleenginepb.Logic_LOGIC_AND, ruleenginepb.Logic_LOGIC_AND},
	}

	ig := NewIngress()

	if _, ok := ig.EvalExpr(newContext(), ruleChain); ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

// nolint
func TestChainVariants_WithMockRequest(t *testing.T) {
	ctx := newContext()

	tests := []struct {
		name   string
		rules  []*ruleenginepb.Rule
		logics []ruleenginepb.Logic
		expect bool
	}{
		{
			name: "AND: path, method, ip all match",
			rules: []*ruleenginepb.Rule{
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_PATH, ruleenginepb.Operator_OPERATOR_EQ, "/v1/login"),
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_METHOD, ruleenginepb.Operator_OPERATOR_EQ, "POST"),
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_IP, ruleenginepb.Operator_OPERATOR_EQ, "203.0.113.42"),
			},
			logics: []ruleenginepb.Logic{
				ruleenginepb.Logic_LOGIC_AND,
				ruleenginepb.Logic_LOGIC_AND,
			},
			expect: true,
		},
		// {
		// 	name: "AND: header mismatch should fail",
		// 	rules: []*ruleenginepb.Rule{
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_HEADER, ruleenginepb.Operator_OPERATOR_EQ, "wrong-header"),
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_PATH, ruleenginepb.Operator_OPERATOR_EQ, "/v1/login"),
		// 	},
		// 	logics: []ruleenginepb.Logic{
		// 		ruleenginepb.Logic_LOGIC_AND,
		// 	},
		// 	expect: false,
		// },
		// {
		// 	name: "OR: header match or path mismatch",
		// 	rules: []*ruleenginepb.Rule{
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_HEADER, ruleenginepb.Operator_OPERATOR_IN, "User-Agent"),
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_PATH, ruleenginepb.Operator_OPERATOR_EQ, "/v1/wrong"),
		// 	},
		// 	logics: []ruleenginepb.Logic{
		// 		ruleenginepb.Logic_LOGIC_OR,
		// 	},
		// 	expect: true,
		// },
		{
			name: "AND: query parameter exists",
			rules: []*ruleenginepb.Rule{
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_QUERY, ruleenginepb.Operator_OPERATOR_IN, "lang"),
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_QUERY, ruleenginepb.Operator_OPERATOR_IN, "lang2"),
			},
			logics: []ruleenginepb.Logic{
				ruleenginepb.Logic_LOGIC_AND,
			},
			expect: true,
		},
		{
			name: "OR: host mismatch but IP match",
			rules: []*ruleenginepb.Rule{
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_HOST, ruleenginepb.Operator_OPERATOR_EQ, "fake.example.com"),
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_IP, ruleenginepb.Operator_OPERATOR_EQ, "203.0.113.42"),
			},
			logics: []ruleenginepb.Logic{
				ruleenginepb.Logic_LOGIC_OR,
			},
			expect: true,
		},
		// {
		// 	name: "AND: body content mismatch should fail",
		// 	rules: []*ruleenginepb.Rule{
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_BODY, ruleenginepb.Operator_OPERATOR_CONTAINS, "john"),
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_BODY, ruleenginepb.Operator_OPERATOR_CONTAINS, "123456"),
		// 	},
		// 	logics: []ruleenginepb.Logic{
		// 		ruleenginepb.Logic_LOGIC_AND,
		// 	},
		// 	expect: false,
		// },
		// {
		// 	name: "OR: body username matches or ip mismatch",
		// 	rules: []*ruleenginepb.Rule{
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_BODY, ruleenginepb.Operator_OPERATOR_CONTAINS, "hung"),
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_IP, ruleenginepb.Operator_OPERATOR_EQ, "198.51.100.10"),
		// 	},
		// 	logics: []ruleenginepb.Logic{
		// 		ruleenginepb.Logic_LOGIC_OR,
		// 	},
		// 	expect: true,
		// },
		// {
		// 	name: "AND: host and TLS must both be true",
		// 	rules: []*ruleenginepb.Rule{
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_HOST, ruleenginepb.Operator_OPERATOR_EQ, "api.example.com"),
		// 		newRule(ruleenginepb.FieldSource_FIELD_SOURCE_TLS, ruleenginepb.Operator_OPERATOR_EQ, "true"),
		// 	},
		// 	logics: []ruleenginepb.Logic{
		// 		ruleenginepb.Logic_LOGIC_AND,
		// 	},
		// 	expect: true,
		// },
		{
			name: "OR: wrong method but correct path",
			rules: []*ruleenginepb.Rule{
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_METHOD, ruleenginepb.Operator_OPERATOR_EQ, "GET"),
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_PATH, ruleenginepb.Operator_OPERATOR_EQ, "/v1/login"),
			},
			logics: []ruleenginepb.Logic{
				ruleenginepb.Logic_LOGIC_OR,
			},
			expect: true,
		},
		{
			name: "AND: wrong IP should fail",
			rules: []*ruleenginepb.Rule{
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_IP, ruleenginepb.Operator_OPERATOR_EQ, structpb.NewStringValue("198.51.100.10")),
				newRule(ruleenginepb.FieldSource_FIELD_SOURCE_PATH, ruleenginepb.Operator_OPERATOR_EQ, "/v1/login"),
			},
			logics: []ruleenginepb.Logic{
				ruleenginepb.Logic_LOGIC_AND,
			},
			expect: false,
		},
	}

	ig := NewIngress()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := &ruleenginepb.Expr{
				Id:      uuid.New().String(),
				Enabled: true,
				Rules:   tt.rules,
				Logics:  tt.logics,
			}
			_, ok := ig.EvalExpr(ctx, chain)
			if ok != tt.expect {
				val, _ := json.Marshal(tt.rules)
				t.Logf("[request][rule] %v", string(val))
				t.Errorf("expected %v, got %v", tt.expect, ok)
			} else {
				t.Logf("%s: passed", tt.name)
			}
		})
	}
}

// nolint
func newRule(src ruleenginepb.FieldSource, op ruleenginepb.Operator, val any) *ruleenginepb.Rule {
	v, _ := structpb.NewValue(val)
	return &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   src,
			Operator: op,
			Value:    v,
			Key:      fmt.Sprintf("%v", val),
		},
	}
}
