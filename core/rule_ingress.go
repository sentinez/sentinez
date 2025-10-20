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

package core

import (
	ruleengpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
)

var _ RuleIngress = (*ruleIngress)(nil)

type RuleIngress interface {
	ExecRule(ctx RequestContext, rule *ruleengpb.Rule) bool
	ExecRuleSet(ctx RequestContext, rule *ruleengpb.RuleSet) bool
}

func NewRuleIngress() RuleIngress {
	return &ruleIngress{}
}

type ruleIngress struct{}

func (ri *ruleIngress) ExecRule(
	ctx RequestContext, rule *ruleengpb.Rule) bool {

	if !rule.GetEnabled() {
		return false
	}

	cond := newCondition(rule.GetCondition())
	ruleCtx := newEvaluator(ctx)
	defer ruleCtx.Release()

	return cond.Accept(ruleCtx)
}

func (ri *ruleIngress) ExecRuleSet(
	ctx RequestContext, ruleSet *ruleengpb.RuleSet) bool {
	if !ruleSet.GetEnabled() {
		return false
	}

	for _, rule := range ruleSet.Rules {
		cond := newCondition(rule.GetCondition())
		ruleCtx := newEvaluator(ctx)

		if !cond.Accept(ruleCtx) {
			return false
		}
	}

	return true
}
