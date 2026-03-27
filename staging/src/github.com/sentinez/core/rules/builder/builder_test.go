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

package builder

import (
	"testing"

	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
)

func newTestRule(id string) *RuleBuilder {
	return NewRule(id, "Test rule").
		Description("Test rule description").
		Priority(100).
		Condition(
			rulepb.FieldSource_FIELD_SOURCE_IP,
			rulepb.Operator_OPERATOR_EQ,
			"",
			"192.168.1.100",
		).
		Action(
			"act-1",
			rulepb.ActionType_ACTION_TYPE_BLOCK,
			map[string]any{"status": 403, "msg": "blocked"},
		)
}

func TestRuleBuilder_JSONExportImport(t *testing.T) {
	ruleBuilder := newTestRule("rule-1")
	rule := ruleBuilder.Build()

	b, err := ruleBuilder.ToJSON()
	if err != nil {
		t.Fatalf("failed to marshal rule to json: %v", err)
	}

	importedBuilder, err := RuleFromJSON(b)
	if err != nil {
		t.Fatalf("failed to unmarshal rule from json: %v", err)
	}

	importedRule := importedBuilder.Build()
	if importedRule.GetId() != rule.GetId() {
		t.Errorf("expected id %v, got %v", rule.GetId(), importedRule.GetId())
	}
	if importedRule.GetCondition().GetSource() != rule.GetCondition().GetSource() {
		t.Errorf("expected source %v, got %v", rule.GetCondition().GetSource(),
			importedRule.GetCondition().GetSource())
	}
	if len(importedRule.GetActions()) != 1 {
		t.Fatalf("expected 1 action, got %v", len(importedRule.GetActions()))
	}

	msgVal := importedRule.GetActions()[0].GetParams().GetFields()["msg"].GetStringValue()
	if msgVal != "blocked" {
		t.Errorf("expected msg 'blocked', got '%s'", msgVal)
	}
}

func newTestExprBuilder() *ExprBuilder {
	r1 := NewRule("r1", "Check Path").
		Condition(
			rulepb.FieldSource_FIELD_SOURCE_PATH,
			rulepb.Operator_OPERATOR_PREFIX,
			"",
			"/api/v1/",
		).Build()

	r2 := NewRule("r2", "Check Method").
		Condition(
			rulepb.FieldSource_FIELD_SOURCE_METHOD,
			rulepb.Operator_OPERATOR_IN,
			"",
			[]any{"POST", "PUT"},
		).Build()

	return NewExpr("expr-1", "API Write Protection").
		Description("Protects API write endpoints").
		AddRule(r1).
		AddLogicAndRule(rulepb.Logic_LOGIC_AND, r2)
}

func TestExprBuilder_JSONExportImport(t *testing.T) {
	exprBuilder := newTestExprBuilder()
	b, err := exprBuilder.ToJSON()
	if err != nil {
		t.Fatalf("failed to marshal expr to json: %v", err)
	}

	importedExprBuilder, err := ExprFromJSON(b)
	if err != nil {
		t.Fatalf("failed to unmarshal expr from json: %v", err)
	}

	importedExpr := importedExprBuilder.Build()

	if importedExpr.GetId() != "expr-1" {
		t.Errorf("expected id expr-1, got %v", importedExpr.GetId())
	}
	if len(importedExpr.GetRules()) != 2 {
		t.Errorf("expected 2 rules, got %v", len(importedExpr.GetRules()))
	}
	if len(importedExpr.GetLogics()) != 1 {
		t.Errorf("expected 1 logic operator, got %v", len(importedExpr.GetLogics()))
	}
	if importedExpr.GetLogics()[0] != rulepb.Logic_LOGIC_AND {
		t.Errorf("expected LOGIC_AND, got %v", importedExpr.GetLogics()[0])
	}
}
