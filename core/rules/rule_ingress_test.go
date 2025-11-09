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
	"testing"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	corehttp "github.com/sentinez/sentinez/core/http"
	corehttpreq "github.com/sentinez/sentinez/core/http/request"
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
			Logic:    ruleenginepb.Logic_LOGIC_AND,
			Value:    structpb.NewStringValue("/v1/login"),
			Key:      "path",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %s", string(val))

	ok := rule.Exec(newContext(), req)
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
			Logic:    ruleenginepb.Logic_LOGIC_AND,
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

	ok := rule.Exec(newContext(), req)

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
			Logic:    ruleenginepb.Logic_LOGIC_AND,
			Value:    structpb.NewStringValue("203.0.113.42"),
			Key:      "ip",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := rule.Exec(newContext(), req)

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
			Logic:    ruleenginepb.Logic_LOGIC_AND,
			Value:    structpb.NewStringValue("203.0.113.0/24"),
			Key:      "ip",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := rule.Exec(newContext(), req)

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
			Logic:    ruleenginepb.Logic_LOGIC_AND,
			Value:    structpb.NewStringValue("203.1.113.0/24"),
			Key:      "ip",
		},
	}

	val, _ := json.Marshal(req)
	t.Logf("[request][rule] %v", string(val))

	ok := rule.Exec(newContext(), req)

	if ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")
}

//nolint:funlen
func TestChain(t *testing.T) {
	rulePath := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_PATH,
			Operator: ruleenginepb.Operator_OPERATOR_EQ,
			Logic:    ruleenginepb.Logic_LOGIC_AND,
			Value:    structpb.NewStringValue("/v1/login"),
			Key:      "path",
		},
	}

	ruleQuery := &ruleenginepb.Rule{
		Enabled: true,
		Condition: &ruleenginepb.Condition{
			Source:   ruleenginepb.FieldSource_FIELD_SOURCE_QUERY,
			Operator: ruleenginepb.Operator_OPERATOR_IN,
			Logic:    ruleenginepb.Logic_LOGIC_AND,
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
			Logic:    ruleenginepb.Logic_LOGIC_AND,
			Value:    structpb.NewStringValue("203.0.113.42"),
			Key:      "ip",
		},
	}

	ruleChain := &ruleenginepb.Chain{
		Enabled: true,
		Rules:   []*ruleenginepb.Rule{rulePath, ruleQuery, ruleClientIP},
	}

	ig := NewIngress()

	if ok := ig.ExecChain(newContext(), ruleChain); ok {
		t.Logf("rule engine matched !!!")
		return
	}

	t.Error("rule engine does not match !!!")

}
