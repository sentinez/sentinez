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
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
	"github.com/sentinez/shared/ids"
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

func toRuleBased(rgLite *ruleenginepb.RuleBasedLite) *ruleenginepb.RuleBased {
	if rgLite == nil {
		return nil
	}

	return &ruleenginepb.RuleBased{
		Id:          rgLite.GetId(),
		Name:        rgLite.GetName(),
		Description: rgLite.GetDescription(),
		Node:        toNode(rgLite.GetNode()),
		Action:      rgLite.GetAction(),
	}
}

func toNode(
	nodeLite *ruleenginepb.RuleBasedLite_NodeLite,
) *ruleenginepb.RuleBased_Node {
	if nodeLite == nil {
		return nil
	}

	node := &ruleenginepb.RuleBased_Node{
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

func toLogic(logic string) ruleenginepb.Logic {
	l, ok := ruleenginepb.Logic_value[logic]
	if !ok {
		return ruleenginepb.Logic_LOGIC_UNSPECIFIED
	}

	return ruleenginepb.Logic(l)
}

func toOperator(operator string) ruleenginepb.Operator {
	op, ok := ruleenginepb.Operator_value[operator]
	if !ok {
		return ruleenginepb.Operator_OPERATOR_UNSPECIFIED
	}

	return ruleenginepb.Operator(op)
}

func toSource(source string) ruleenginepb.FieldSource {
	src, ok := ruleenginepb.FieldSource_value[source]
	if !ok {
		return ruleenginepb.FieldSource_FIELD_SOURCE_UNSPECIFIED
	}

	return ruleenginepb.FieldSource(src)
}

func toCondition(cond *ruleenginepb.ConditionLite) *ruleenginepb.Condition {
	if cond == nil {
		return nil
	}

	return &ruleenginepb.Condition{
		Id:       ids.NewNanoID(condPrefix),
		Key:      cond.GetKey(),
		Operator: toOperator(cond.GetOperator()),
		Value:    structpb.NewStringValue(cond.GetValue()),
		Source:   toSource(cond.GetSource()),
	}
}

func toRule(rule *ruleenginepb.RuleLite) *ruleenginepb.Rule {
	if rule == nil {
		return nil
	}

	return &ruleenginepb.Rule{
		Id:        ids.NewNanoID(rulePrefix),
		Name:      rule.GetName(),
		Condition: toCondition(rule.GetCondition()),
	}
}
