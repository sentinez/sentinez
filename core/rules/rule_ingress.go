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
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	corehttp "github.com/sentinez/sentinez/core/http"
	"github.com/sentinez/sentinez/core/internal/zlog"
)

var _ Rules = (*ingress)(nil)

type Rules interface {
	Exec(ctx corehttp.RequestContext, rule *rulepb.Rule) bool
	ExecChain(ctx corehttp.RequestContext, rule *rulepb.Chain) bool
}

func NewIngress() Rules {
	return &ingress{}
}

type ingress struct{}

func (i *ingress) Exec(ctx corehttp.RequestContext, rule *rulepb.Rule) bool {

	if !rule.GetEnabled() {
		return false
	}

	cond := newCondition(rule.GetCondition())
	ruleCtx := newEvaluator(ctx)
	defer ruleCtx.Release()

	return cond.Accept(ruleCtx)
}

// ExecChain a list of rule
//
// eg:
// A OR B AND C => A OR (B AND C)
// run A, if A is true, return true and stop chain, if A is false run B AND C
// run B, if B is true, run C, else if B is false, stop the chain, not run C
func (i *ingress) ExecChain(
	ctx corehttp.RequestContext, chain *rulepb.Chain) bool {

	if !chain.GetEnabled() {
		return false
	}

	result := true

	for i, rule := range chain.Rules {

		cond := newCondition(rule.GetCondition())
		ruleCtx := newEvaluator(ctx)

		switch rule.GetCondition().GetLogic() {
		case rulepb.Logic_LOGIC_OR:
			// with logic OR, we use `||` to combine all results
			// first element in rule array, or next accept is fasle
			// we accept next condition util last element
			// if result of the accept is true, return true
			result = result || cond.Accept(ruleCtx)
			zlog.Debugf("[index=%d] return %v op=%s", i, cond.Accept(ruleCtx), rule.GetCondition().GetLogic())

			if result {
				return true
			}

		default:
			// with logic AND, we use `&&` to combine all results
			// if first element in array, or next accept is false
			// we stop the chain, and return false
			// (its mean chains of rule are not match)
			result = result && cond.Accept(ruleCtx)
			zlog.Debugf("[index=%d] return %v op=%s", i, cond.Accept(ruleCtx), rule.GetCondition().GetLogic())
		}
	}

	return result
}
