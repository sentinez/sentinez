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

package ruleengine

import (
	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	rulectx "github.com/sentinez/sentinez/corerule/context"
	"github.com/sentinez/sentinez/corerule/engine/condition"
)

var _ Ingress = (*ingress)(nil)

type Ingress interface {
	ExecRule(ctx rulectx.Context, rule *ruleenginepb.Rule) bool
	ExecRuleSet(ctx rulectx.Context, rule *ruleenginepb.RuleSet) bool
}

func NewIngress() Ingress {
	return &ingress{}
}

type ingress struct{}

func (i *ingress) ExecRule(ctx rulectx.Context, rule *ruleenginepb.Rule) bool {
	if !rule.GetEnabled() {
		return false
	}

	cond := condition.New(rule.GetCondition())
	ruleCtx := condition.NewEvaluator(ctx)
	defer ruleCtx.Release()

	return cond.Accept(ruleCtx)
}

func (i *ingress) ExecRuleSet(
	ctx rulectx.Context, ruleSet *ruleenginepb.RuleSet) bool {
	if !ruleSet.GetEnabled() {
		return false
	}

	for _, rule := range ruleSet.Rules {
		cond := condition.New(rule.GetCondition())
		ruleCtx := condition.NewEvaluator(ctx)

		if !cond.Accept(ruleCtx) {
			return false
		}
	}

	return true
}
