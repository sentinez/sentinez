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

	corehttp "github.com/sentinez/core/http"
	corehttpreq "github.com/sentinez/core/http/request"
	httppb "github.com/sentinez/sentinez/api/gen/go/sentinez/network/http/v1"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/structpb"
)

// nolint
func newBaseContext() *httppb.Request {
	return &httppb.Request{
		Body: []byte(`{"username":"hung","password":"123456"}`),
		Headers: []*httppb.RequestHeader{
			{Key: []byte("Content-Type"), Values: [][]byte{[]byte("application/json")}},
			{Key: []byte("User-Agent"), Values: [][]byte{[]byte("curl/8.0.1")}},
			{Key: []byte("Accept"), Values: [][]byte{[]byte("*/*")}},
			{Key: []byte("Authorization"), Values: [][]byte{[]byte("Bearer abc.def.ghi")}},
		},
		Host:        "api.example.com",
		ClientIp:    "203.0.113.42",
		Fingerprint: "ja4:abcd1234efgh5678ijkl9012mnop3456",
		Method:      "POST",
		Path:        []byte("/v1/login"),
		Queries: []*httppb.RequestQuery{
			{Key: []byte("lang"), Values: [][]byte{[]byte("vi")}},
			{Key: []byte("lang2"), Values: [][]byte{[]byte("vi")}},
		},
		Scheme:        "https",
		Protocol:      "HTTP/1.1",
		RemoteAddress: "203.0.113.42:52341",
		Status:        200,
		Uri:           []byte("https://api.example.com/v1/login?redirect=/home&lang=en"),
	}
}

//nolint:lll
func newContext() corehttp.RequestContext {
	reqCtx := newBaseContext()
	return corehttpreq.NewRequestContext(context.Background(), reqCtx)
}

func TestRulePath(t *testing.T) {
	req := &rulepb.Rule{
		Condition: &rulepb.Condition{
			Source:   rulepb.FieldSource_FIELD_SOURCE_PATH,
			Operator: rulepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("/v1/login"),
			Key:      "path",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %s", string(val))

	ok := eval(newContext(), req)
	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

func TestRuleQuery(t *testing.T) {
	req := &rulepb.Rule{
		Condition: &rulepb.Condition{
			Source:   rulepb.FieldSource_FIELD_SOURCE_QUERY,
			Operator: rulepb.Operator_OPERATOR_IN,
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

	ok := eval(newContext(), req)

	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

func TestRuleClientIP(t *testing.T) {
	req := &rulepb.Rule{
		Condition: &rulepb.Condition{
			Source:   rulepb.FieldSource_FIELD_SOURCE_IP,
			Operator: rulepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("203.0.113.42"),
			Key:      "ip",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := eval(newContext(), req)

	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

func TestRuleClientIPRange(t *testing.T) {
	req := &rulepb.Rule{
		Condition: &rulepb.Condition{
			Source:   rulepb.FieldSource_FIELD_SOURCE_IP,
			Operator: rulepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("203.0.113.0/24"),
			Key:      "ip",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := eval(newContext(), req)

	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

func TestRuleClientIPRangeNotEQ(t *testing.T) {

	req := &rulepb.Rule{
		Condition: &rulepb.Condition{
			Source:   rulepb.FieldSource_FIELD_SOURCE_IP,
			Operator: rulepb.Operator_OPERATOR_NE,
			Value:    structpb.NewStringValue("203.1.113.0/24"),
			Key:      "ip",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := eval(newContext(), req)

	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

// nolint
func TestChain(t *testing.T) {
	// Directly build RuleBased using newRule helper
	rg := &rulepb.RuleBased{
		Node: &rulepb.RuleBased_Node{
			Operator: rulepb.Logic_LOGIC_AND,
			Rules: []*rulepb.Rule{
				newRule(rulepb.FieldSource_FIELD_SOURCE_PATH, rulepb.Operator_OPERATOR_EQ, "/v1/login"),
				newRule(rulepb.FieldSource_FIELD_SOURCE_QUERY, rulepb.Operator_OPERATOR_IN, []any{"lang"}), // Corrected for existence check
				newRule(rulepb.FieldSource_FIELD_SOURCE_IP, rulepb.Operator_OPERATOR_EQ, "203.0.113.42"),
			},
		},
	}

	ig := NewIngress(rg)
	matched := &rulepb.MatchedRules{}
	if ok := ig.Eval(newContext(), matched); ok {
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
		rg     *rulepb.RuleBased
		expect bool
	}{
		{
			name: "AND: path, method, ip all match",
			rg: &rulepb.RuleBased{
				Node: &rulepb.RuleBased_Node{
					Operator: rulepb.Logic_LOGIC_AND,
					Rules: []*rulepb.Rule{
						newRule(rulepb.FieldSource_FIELD_SOURCE_PATH, rulepb.Operator_OPERATOR_EQ, "/v1/login"),
						newRule(rulepb.FieldSource_FIELD_SOURCE_METHOD, rulepb.Operator_OPERATOR_EQ, "POST"),
						newRule(rulepb.FieldSource_FIELD_SOURCE_IP, rulepb.Operator_OPERATOR_EQ, "203.0.113.42"),
					},
				},
			},
			expect: true,
		},
		{
			name: "OR: host mismatch but IP match",
			rg: &rulepb.RuleBased{
				Node: &rulepb.RuleBased_Node{
					Operator: rulepb.Logic_LOGIC_OR,
					Rules: []*rulepb.Rule{
						newRule(rulepb.FieldSource_FIELD_SOURCE_HOST, rulepb.Operator_OPERATOR_EQ, "fake.example.com"),
						newRule(rulepb.FieldSource_FIELD_SOURCE_IP, rulepb.Operator_OPERATOR_EQ, "203.0.113.42"),
					},
				},
			},
			expect: true,
		},
		{
			name: "NESTED: A AND (B OR C)",
			// A (IP match), B (Path mismatch), C (Method match) -> True
			rg: &rulepb.RuleBased{
				Node: &rulepb.RuleBased_Node{
					Operator: rulepb.Logic_LOGIC_AND,
					Rules: []*rulepb.Rule{
						newRule(rulepb.FieldSource_FIELD_SOURCE_IP, rulepb.Operator_OPERATOR_EQ, "203.0.113.42"), // A
					},
					Groups: []*rulepb.RuleBased_Node{
						{
							Operator: rulepb.Logic_LOGIC_OR,
							Rules: []*rulepb.Rule{
								newRule(rulepb.FieldSource_FIELD_SOURCE_PATH, rulepb.Operator_OPERATOR_EQ, "/wrong"), // B
								newRule(rulepb.FieldSource_FIELD_SOURCE_METHOD, rulepb.Operator_OPERATOR_EQ, "POST"), // C
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "NOT: NOT (Method GET)",
			// Method is POST -> NOT (POST == GET) -> NOT (false) -> True
			rg: &rulepb.RuleBased{
				Node: &rulepb.RuleBased_Node{
					Operator: rulepb.Logic_LOGIC_NOT,
					Rules: []*rulepb.Rule{
						newRule(rulepb.FieldSource_FIELD_SOURCE_METHOD, rulepb.Operator_OPERATOR_EQ, "GET"),
					},
				},
			},
			expect: true,
		},
	}

	matched := &rulepb.MatchedRules{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ig := NewIngress(tt.rg)
			if ok := ig.Eval(ctx, matched); ok != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, ok)
			} else {
				t.Logf("%s: passed", tt.name)
			}
		})
	}
}

// nolint
func newRule(src rulepb.FieldSource, op rulepb.Operator, val any) *rulepb.Rule {
	// structpb.NewValue handles []interface{} as ListValue.
	// We convert common slice types to []interface{} to ensure correct behavior.
	var finalVal any = val
	switch v := val.(type) {
	case []string:
		items := make([]any, len(v))
		for i, s := range v {
			items[i] = s
		}
		finalVal = items
	}

	v, err := structpb.NewValue(finalVal)
	if err != nil {
		zlog.Warnf("rule: new value error: %v", err)
	}

	key := ""
	if src != rulepb.FieldSource_FIELD_SOURCE_QUERY && src != rulepb.FieldSource_FIELD_SOURCE_HEADER {
		key = fmt.Sprintf("%v", val)
	}

	return &rulepb.Rule{
		Condition: &rulepb.Condition{
			Source:   src,
			Operator: op,
			Value:    v,
			Key:      key,
		},
	}
}

// nolint
func newRuleValue(src rulepb.FieldSource, op rulepb.Operator, val *structpb.Value) *rulepb.Rule {
	return &rulepb.Rule{
		Condition: &rulepb.Condition{
			Source:   src,
			Operator: op,
			Value:    val,
			Key:      fmt.Sprintf("%v", val),
		},
	}
}

func BenchmarkEvalRule(b *testing.B) {
	zlog.SetLogLevel(zlog.LevelFatal)

	req := &rulepb.Rule{
		Condition: &rulepb.Condition{
			Source:   rulepb.FieldSource_FIELD_SOURCE_PATH,
			Operator: rulepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("/v1/login"),
			Key:      "path",
		},
	}
	ctx := newContext()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eval(ctx, req)
	}
}

// nolint
func BenchmarkEvalRuleBased_Simple(b *testing.B) {
	zlog.SetLogLevel(zlog.LevelFatal)

	rg := &rulepb.RuleBased{
		Node: &rulepb.RuleBased_Node{
			Operator: rulepb.Logic_LOGIC_AND,
			Rules: []*rulepb.Rule{
				newRule(rulepb.FieldSource_FIELD_SOURCE_PATH, rulepb.Operator_OPERATOR_EQ, "/v1/login"),
				newRule(rulepb.FieldSource_FIELD_SOURCE_METHOD, rulepb.Operator_OPERATOR_EQ, "POST"),
				newRule(rulepb.FieldSource_FIELD_SOURCE_IP, rulepb.Operator_OPERATOR_EQ, "203.0.113.42"),
			},
		},
	}
	ig := NewIngress(rg)

	ctx := newContext()
	matched := &rulepb.MatchedRules{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ig.Eval(ctx, matched)
	}
}

// nolint
func BenchmarkEvalRuleBased_Complex(b *testing.B) {
	zlog.SetLogLevel(zlog.LevelFatal)

	rg := &rulepb.RuleBased{
		Node: &rulepb.RuleBased_Node{
			Operator: rulepb.Logic_LOGIC_AND,
			Rules: []*rulepb.Rule{
				newRule(rulepb.FieldSource_FIELD_SOURCE_IP, rulepb.Operator_OPERATOR_EQ, "203.0.113.42"),
			},
			Groups: []*rulepb.RuleBased_Node{
				{
					Operator: rulepb.Logic_LOGIC_OR,
					Rules: []*rulepb.Rule{
						newRule(rulepb.FieldSource_FIELD_SOURCE_PATH, rulepb.Operator_OPERATOR_EQ, "/wrong"),
						newRule(rulepb.FieldSource_FIELD_SOURCE_METHOD, rulepb.Operator_OPERATOR_EQ, "POST"),
					},
				},
				{
					Operator: rulepb.Logic_LOGIC_NOT,
					Rules: []*rulepb.Rule{
						newRule(rulepb.FieldSource_FIELD_SOURCE_METHOD, rulepb.Operator_OPERATOR_EQ, "GET"),
					},
				},
			},
		},
	}

	ig := NewIngress(rg)
	ctx := newContext()
	matched := &rulepb.MatchedRules{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ig.Eval(ctx, matched)
	}
}

// nolint
func BenchmarkEvalRuleBased_Complex_Parallel(b *testing.B) {
	zlog.SetLogLevel(zlog.LevelFatal)

	rg := &rulepb.RuleBased{
		Node: &rulepb.RuleBased_Node{
			Operator: rulepb.Logic_LOGIC_AND,
			Rules: []*rulepb.Rule{
				newRule(rulepb.FieldSource_FIELD_SOURCE_IP, rulepb.Operator_OPERATOR_EQ, "203.0.113.42"),
			},
			Groups: []*rulepb.RuleBased_Node{
				{
					Operator: rulepb.Logic_LOGIC_OR,
					Rules: []*rulepb.Rule{
						newRule(rulepb.FieldSource_FIELD_SOURCE_PATH, rulepb.Operator_OPERATOR_EQ, "/wrong"),
						newRule(rulepb.FieldSource_FIELD_SOURCE_METHOD, rulepb.Operator_OPERATOR_EQ, "POST"),
					},
				},
				{
					Operator: rulepb.Logic_LOGIC_NOT,
					Rules: []*rulepb.Rule{
						newRule(rulepb.FieldSource_FIELD_SOURCE_METHOD, rulepb.Operator_OPERATOR_EQ, "GET"),
					},
				},
			},
		},
	}

	ig := NewIngress(rg)

	ctx := newContext()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = ig.Eval(ctx, nil)
		}
	})
}
