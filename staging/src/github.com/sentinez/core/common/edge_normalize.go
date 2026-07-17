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

package corecmn

import (
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
	"github.com/sentinez/shared/rand"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	rgPrefix   = "senz.rulebased."
	condPrefix = "senz.cond."
	rulePrefix = "senz.rule."
)

func NormalizeEdgeSetting(edge *edgepb.Setting) {
	normalizeEdgeSecurity(edge.GetSecurity())
}

func normalizeEdgeSecurity(edgeSec *edgepb.Security) {
	rgLite := edgeSec.GetRuleBased()
	if rgLite == nil {
		return
	}

	edgeSec.RuleBasedCompiled = toRuleBased(rgLite)
}

func toRuleBased(rgLite *rulepb.RuleBasedLite) *rulepb.RuleBased {
	if rgLite == nil {
		return nil
	}

	return &rulepb.RuleBased{
		Id:          rgLite.GetId(),
		Name:        rgLite.GetName(),
		Description: rgLite.GetDescription(),
		Node:        toNode(rgLite.GetNode()),
		Action:      rgLite.GetAction(),
	}
}

func toNode(
	nodeLite *rulepb.RuleBasedLite_NodeLite,
) *rulepb.RuleBased_Node {
	if nodeLite == nil {
		return nil
	}

	node := &rulepb.RuleBased_Node{
		Operator: toLogic(nodeLite.GetOperator()),
	}

	for _, rLite := range nodeLite.GetRules() {
		node.Rules = append(node.Rules, toRule(rLite))
	}

	for _, gLite := range nodeLite.GetGroups() {
		node.Groups = append(node.Groups, toNode(gLite))
	}

	return node
}

func toLogic(logic string) rulepb.Logic {
	l, ok := rulepb.Logic_value[logic]
	if !ok {
		return rulepb.Logic_LOGIC_UNSPECIFIED
	}

	return rulepb.Logic(l)
}

func toOperator(operator string) rulepb.Operator {
	op, ok := rulepb.Operator_value[operator]
	if !ok {
		return rulepb.Operator_OPERATOR_UNSPECIFIED
	}

	return rulepb.Operator(op)
}

func toSource(source string) rulepb.FieldSource {
	src, ok := rulepb.FieldSource_value[source]
	if !ok {
		return rulepb.FieldSource_FIELD_SOURCE_UNSPECIFIED
	}

	return rulepb.FieldSource(src)
}

func toCondition(cond *rulepb.ConditionLite) *rulepb.Condition {
	if cond == nil {
		return nil
	}

	return &rulepb.Condition{
		Id:       rand.NewNanoID(condPrefix),
		Key:      cond.GetKey(),
		Operator: toOperator(cond.GetOperator()),
		Value:    structpb.NewStringValue(cond.GetValue()),
		Source:   toSource(cond.GetSource()),
	}
}

func toRule(rule *rulepb.RuleLite) *rulepb.Rule {
	if rule == nil {
		return nil
	}

	return &rulepb.Rule{
		Id:        rand.NewNanoID(rulePrefix),
		Name:      rule.GetName(),
		Condition: toCondition(rule.GetCondition()),
	}
}
