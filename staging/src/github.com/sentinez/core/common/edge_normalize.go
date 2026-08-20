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
	"github.com/sentinez/shared/zlog"
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

	zlog.Debugf("Expr: %v", toExpr(rgLite.GetExpr()))

	return &rulepb.RuleBased{
		Id:          rgLite.GetId(),
		Name:        rgLite.GetName(),
		Description: rgLite.GetDescription(),
		Expr:        toExpr(rgLite.GetExpr()),
		Action:      rgLite.GetAction(),
	}
}

func toExpr(expr *rulepb.ExpressionLite) *rulepb.Expression {
	e := &rulepb.Expression{
		OrCondition: parsOrCond(expr.GetOrCondition()),
	}

	return e
}

func parsOrCond(orCond []*rulepb.AndConditionLite) []*rulepb.AndCondition {
	if orCond == nil {
		return nil
	}

	andCond := []*rulepb.AndCondition{}

	for _, and := range orCond {
		a := parseAndCond(and)
		if a == nil {
			continue
		}

		andCond = append(andCond, a)
	}

	return andCond
}

func parseAndCond(andCond *rulepb.AndConditionLite) *rulepb.AndCondition {
	if andCond == nil {
		return nil
	}

	and := &rulepb.AndCondition{}

	for _, rule := range andCond.GetRules() {
		and.Rules = append(and.Rules, toRule(rule))
	}

	and.OrCondition = parsOrCond(andCond.GetOrCondition())

	return and
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
