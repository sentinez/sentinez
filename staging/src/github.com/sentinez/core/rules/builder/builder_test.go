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

	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
)

// nolint
func TestBuilder(t *testing.T) {
	r1 := NewRule().
		WithID("r1").
		WithName("Path Rule").
		WithCondition(ruleenginepb.FieldSource_FIELD_SOURCE_PATH, ruleenginepb.Operator_OPERATOR_EQ, "/v1/login").
		Build()

	r2 := NewRule().
		WithID("r2").
		WithName("Method Rule").
		WithCondition(ruleenginepb.FieldSource_FIELD_SOURCE_METHOD, ruleenginepb.Operator_OPERATOR_IN, []string{"GET", "POST"}).
		Build()

	subGroup := NewGroup(ruleenginepb.Logic_LOGIC_OR).
		WithName("Sub Group").
		AddRule(r1).
		AddRule(r2).
		Build()

	root := NewGroup(ruleenginepb.Logic_LOGIC_AND).
		WithID("root").
		WithName("Root Group").
		AddRule(NewRule().WithCondition(ruleenginepb.FieldSource_FIELD_SOURCE_IP, ruleenginepb.Operator_OPERATOR_EQ, "127.0.0.1").Build()).
		AddGroup(subGroup).
		Build()

	if root.Id != "root" {
		t.Errorf("expected root ID to be 'root', got %s", root.Id)
	}

	if len(root.Node.Rules) != 1 {
		t.Errorf("expected 1 rule in root, got %d", len(root.Node.Rules))
	}

	if len(root.Node.Groups) != 1 {
		t.Errorf("expected 1 group in root, got %d", len(root.Node.Groups))
	}

	if root.Node.Groups[0].Operator != ruleenginepb.Logic_LOGIC_OR {
		t.Errorf("expected subgroup operator to be OR, got %v", root.Node.Groups[0].Operator)
	}

	if len(root.Node.Groups[0].Rules) != 2 {
		t.Errorf("expected 2 rules in subgroup, got %d", len(root.Node.Groups[0].Rules))
	}
}

// nolint
func TestConvenienceHelpers(t *testing.T) {
	r1 := NewRule().WithID("1").Build()
	r2 := NewRule().WithID("2").Build()

	andGroup := And(r1, r2).Build()
	if andGroup.Node.Operator != ruleenginepb.Logic_LOGIC_AND || len(andGroup.Node.Rules) != 2 {
		t.Error("And helper failed")
	}

	orGroup := Or(r1, r2).Build()
	if orGroup.Node.Operator != ruleenginepb.Logic_LOGIC_OR || len(orGroup.Node.Rules) != 2 {
		t.Error("Or helper failed")
	}

	notGroup := Not(r1).Build()
	if notGroup.Node.Operator != ruleenginepb.Logic_LOGIC_NOT || len(notGroup.Node.Rules) != 1 {
		t.Error("Not helper failed for rule")
	}

	notSubGroup := Not(andGroup).Build()
	if notSubGroup.Node.Operator != ruleenginepb.Logic_LOGIC_NOT || len(notSubGroup.Node.Groups) != 1 {
		t.Error("Not helper failed for group")
	}
}
