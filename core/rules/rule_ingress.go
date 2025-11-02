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

package rules

import (
	ruleengpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	"github.com/sentinez/sentinez/core/networks"
)

var _ Rules = (*ingress)(nil)

type Rules interface {
	Exec(ctx networks.Context, rule *ruleengpb.Rule) bool
	ExecChain(ctx networks.Context, rule *ruleengpb.Chain) bool
}

func NewIngress() Rules {
	return &ingress{}
}

type ingress struct{}

func (i *ingress) Exec(ctx networks.Context, rule *ruleengpb.Rule) bool {

	if !rule.GetEnabled() {
		return false
	}

	cond := newCondition(rule.GetCondition())
	ruleCtx := newEvaluator(ctx)
	defer ruleCtx.Release()

	return cond.Accept(ruleCtx)
}

func (i *ingress) ExecChain(ctx networks.Context, chain *ruleengpb.Chain) bool {

	if !chain.GetEnabled() {
		return false
	}

	for _, rule := range chain.Rules {
		cond := newCondition(rule.GetCondition())
		ruleCtx := newEvaluator(ctx)

		if !cond.Accept(ruleCtx) {
			return false
		}
	}

	return true
}
