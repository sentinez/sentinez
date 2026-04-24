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
	chttp "github.com/sentinez/core/http"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
)

var _ Rules = (*ingress)(nil)

type MatchedFunc func(ctx chttp.RequestContext,
	rule *rulepb.Rule) (id string, name string, ok bool)

// nolint
type Rules interface {
	Eval(ctx chttp.RequestContext, m *rulepb.MatchedRules) bool
	Action() *rulepb.Action
}

func NewIngress(rg *rulepb.RuleBased) Rules {
	root := buildNode(rg.GetNode(), match)

	return &ingress{
		root:   root,
		action: rg.GetAction(),
	}
}

type ingress struct {
	root   *node
	action *rulepb.Action
}

func eval(ctx chttp.RequestContext, rule *rulepb.Rule) bool {
	// zlog.Debugf("[edge][%s] >>> visit ingress eval", ctx.RequestId())
	return accept(ctx, rule.GetCondition())
}

func match(ctx chttp.RequestContext,
	rule *rulepb.Rule) (id string, name string, ok bool) {

	if ok = eval(ctx, rule); !ok {
		return "", "", false
	}

	return rule.GetId(), rule.GetName(), true
}

// Eval a nested group of rules
func (in *ingress) Eval(ctx chttp.RequestContext, m *rulepb.MatchedRules) bool {
	if in.root == nil {
		return false
	}

	if ok := in.root.eval(ctx, m); ok {
		return true
	}

	return false
}

func (in *ingress) Action() *rulepb.Action {
	return in.action
}
