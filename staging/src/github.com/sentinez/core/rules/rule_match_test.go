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
	"testing"

	corehttp "github.com/sentinez/core/http"
	corehttpreq "github.com/sentinez/core/http/request"
	httppb "github.com/sentinez/sentinez/api/gen/go/sentinez/network/http/v1"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// nolint
func newTestContext() corehttp.RequestContext {
	reqCtx := &httppb.Request{
		Body: []byte(`{"message":"hello world"}`),
		Headers: []*httppb.RequestHeader{
			{Key: []byte("Host"), Values: [][]byte{[]byte("example.com")}},
			{Key: []byte("User-Agent"), Values: [][]byte{[]byte("test-agent/1.0")}},
		},
		Host:        "example.com",
		ClientIp:    "192.168.1.100",
		Fingerprint: "t13d1516h2_8daaf1195655_e49aeecd9438",
		Method:      "PUT",
		Path:        []byte("/api/test"),
		Queries: []*httppb.RequestQuery{
			{Key: []byte("search"), Values: [][]byte{[]byte("golang")}},
			{Key: []byte("page"), Values: [][]byte{[]byte("1")}},
		},
		Scheme: "https",
	}
	return corehttpreq.NewRequestContext(context.Background(), reqCtx)
}

// nolint
func TestMatchSourceBody(t *testing.T) {
	ctx := newTestContext()

	tests := []struct {
		name   string
		op     rulepb.Operator
		val    string
		expect bool
	}{
		{"Contains hello", rulepb.Operator_OPERATOR_CONTAINS, "hello", true},
		{"Contains foo", rulepb.Operator_OPERATOR_CONTAINS, "foo", false},
		{"Prefix {", rulepb.Operator_OPERATOR_PREFIX, "{", true},
		{"Suffix }", rulepb.Operator_OPERATOR_SUFFIX, "}", true},
		{"Matches regex", rulepb.Operator_OPERATOR_MATCHES, "^\\{.*\\}$", true},
		{"Matches regex fail", rulepb.Operator_OPERATOR_MATCHES, "^foo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := newRule(rulepb.FieldSource_FIELD_SOURCE_BODY, tt.op, tt.val)
			if matchSourceBody(ctx, rule.Condition) != tt.expect {
				t.Errorf("expected %v", tt.expect)
			}
		})
	}
}

// nolint
func TestMatchSourceJA4(t *testing.T) {
	ctx := newTestContext()

	tests := []struct {
		name   string
		op     rulepb.Operator
		val    string
		expect bool
	}{
		{"EQ exact", rulepb.Operator_OPERATOR_EQ, "t13d1516h2_8daaf1195655_e49aeecd9438", true},
		{"EQ wrong", rulepb.Operator_OPERATOR_EQ, "t13", false},
		{"PREFIX t13", rulepb.Operator_OPERATOR_PREFIX, "t13", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := newRule(rulepb.FieldSource_FIELD_SOURCE_JA4, tt.op, tt.val)
			if matchSourceJA4(ctx, rule.Condition) != tt.expect {
				t.Errorf("expected %v", tt.expect)
			}
		})
	}
}

// nolint
func TestMatchSourceIPList(t *testing.T) {
	ctx := newTestContext() // IP is 192.168.1.100

	listVal := structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{
		structpb.NewStringValue("10.0.0.0/8"),
		structpb.NewStringValue("192.168.1.0/24"),
	}})

	ruleIn := newRuleValue(rulepb.FieldSource_FIELD_SOURCE_IP, rulepb.Operator_OPERATOR_IN, listVal)
	if matchSourceIP(ctx, ruleIn.Condition) != true {
		t.Errorf("expected true for IN")
	}

	ruleNotIn := newRuleValue(rulepb.FieldSource_FIELD_SOURCE_IP, rulepb.Operator_OPERATOR_NOT_IN, listVal)
	if matchSourceIP(ctx, ruleNotIn.Condition) != false {
		t.Errorf("expected false for NOT_IN")
	}
}

// nolint
func TestMatchSourceHeaderKey(t *testing.T) {
	ctx := newTestContext() // User-Agent: test-agent/1.0

	rule := newRule(rulepb.FieldSource_FIELD_SOURCE_HEADER, rulepb.Operator_OPERATOR_CONTAINS, "test-agent")
	rule.Condition.Key = "User-Agent"

	if matchSourceHeader(ctx, rule.Condition) != true {
		t.Errorf("expected true")
	}
}

// nolint
func TestMatchSourceMethodList(t *testing.T) {
	ctx := newTestContext() // Method is PUT

	listVal := structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{
		structpb.NewStringValue("POST"),
		structpb.NewStringValue("PUT"),
	}})

	ruleIn := newRuleValue(rulepb.FieldSource_FIELD_SOURCE_METHOD, rulepb.Operator_OPERATOR_IN, listVal)
	if matchSourceMethod(ctx, ruleIn.Condition) != true {
		t.Errorf("expected true for IN PUT")
	}
}

// nolint
func TestMatchSourceTLS(t *testing.T) {
	ctx := newTestContext() // TLS is true

	ruleTlsTrue := newRule(rulepb.FieldSource_FIELD_SOURCE_TLS, rulepb.Operator_OPERATOR_EQ, "true")
	if matchSourceTLS(ctx, ruleTlsTrue.Condition) != true {
		t.Errorf("expected true for matching true")
	}

	ruleTlsFalse := newRule(rulepb.FieldSource_FIELD_SOURCE_TLS, rulepb.Operator_OPERATOR_EQ, "false")
	if matchSourceTLS(ctx, ruleTlsFalse.Condition) != false {
		t.Errorf("expected false for matching false")
	}
}
