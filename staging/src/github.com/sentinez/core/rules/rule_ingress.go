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
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	"github.com/sentinez/shared/sync"
)

var _ Rules = (*ingress)(nil)

type MatchedFunc func(ctx chttp.RequestContext,
	rule *rulepb.Rule) (id string, name string, ok bool)

type Rules interface {
	Eval(ctx chttp.RequestContext, rule *rulepb.Rule) bool
	EvalExpr(ctx chttp.RequestContext,
		rule *rulepb.Expr) (*rulepb.MatchedRules, bool)
}

func NewIngress() Rules {
	return &ingress{}
}

type ingress struct {
	expr sync.Map[string, *exprs]
}

func (in *ingress) Eval(ctx chttp.RequestContext, rule *rulepb.Rule) bool {
	// zlog.Debugf("[edge][%s] >>> visit ingress eval", ctx.RequestId())

	if !rule.GetEnabled() {
		return false
	}

	cond := newCondition(rule.GetCondition())
	ruleCtx := newEvaluator(ctx)
	defer ruleCtx.Release()

	return cond.Accept(ruleCtx)
}

func (in *ingress) matched(ctx chttp.RequestContext,
	rule *rulepb.Rule) (id string, name string, ok bool) {

	if ok = in.Eval(ctx, rule); !ok {
		return "", "", false
	}

	return rule.GetId(), rule.GetName(), true
}

// EvalExpr a list of rule
func (in *ingress) EvalExpr(
	ctx chttp.RequestContext, chain *rulepb.Expr) (*rulepb.MatchedRules, bool) {
	// zlog.Debugf("[edge][%s] >>> visit ingress", ctx.RequestId())

	if !chain.GetEnabled() {
		return nil, false
	}

	expr, ok := in.expr.Load(chain.GetId())
	if !ok {
		expr = newExpr(chain)
		in.expr.Store(chain.GetId(), expr)
	}

	expr.tx.reset()

	if ok = expr.build(in.matched).eval(ctx); ok {
		return expr.tx.matched, ok
	}

	return nil, false
}
