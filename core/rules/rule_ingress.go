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
	"sync"

	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	chttp "github.com/sentinez/sentinez/core/http"
	"github.com/sentinez/sentinez/core/internal/logic"
	"github.com/sentinez/sentinez/shared/zlog"
)

var _ Rules = (*ingress)(nil)

type MatchedFunc func(ctx chttp.RequestContext,
	rule *rulepb.Rule) (id string, name string, score int32, ok bool)

type Rules interface {
	Eval(ctx chttp.RequestContext, rule *rulepb.Rule) bool
	EvalExpr(ctx chttp.RequestContext,
		rule *rulepb.Expr) (*rulepb.MatchedRules, bool)
}

type tx struct {
	matched    *rulepb.MatchedRules
	threshold  int32
	totalScore int32
}

type exprs struct {
	chain *rulepb.Expr
	tx    *tx
}

// Example context:
//
// rules:  a b c d
// logics: OR AND AND
// expr:   a OR b AND c AND d
//
// The expression tree built from this looks like:
//
//	  OR
//	 /  \
//	a    AND
//	    /   \
//	   b     AND
//	        /   \
//	       c     d
//
// So the evaluation order (with short-circuiting) becomes:
//
//	a || (b && c && d)
//
//nolint:funlen
func (ex *exprs) build(exec MatchedFunc) *logic.Node {

	// Get the list of rules and logical operators (AND / OR)
	rules := ex.chain.GetRules()
	logics := ex.chain.GetLogics()

	// Sanity check: in a valid expression,
	// the number of logics = number of rules - 1
	if len(logics)+1 != len(rules) {
		return nil
	}

	// Step 1: Convert all rules into leaf nodes in the logic tree.
	// Each rule becomes a node that can execute and return true/false.
	nodes := make([]*logic.Node, len(rules))
	for i, r := range rules {
		// idx := i // capture index for logging
		nodes[i] = logic.NewNode(func(ctx chttp.RequestContext) bool {

			id, name, score, ok := exec(ctx, r)
			if ok {
				ex.tx.matched.Ids = append(ex.tx.matched.Ids, id)
				ex.tx.matched.Names = append(ex.tx.matched.Names, name)
				ex.tx.matched.Scores = append(ex.tx.matched.Scores, score)
				ex.tx.totalScore += score
			}

			if ex.tx.threshold != 0 && ex.tx.totalScore >= ex.tx.threshold {
				return true
			}

			return ok
		})
	}

	// Step 2: Start building the tree from the leftmost node.
	current := nodes[0]
	i := 1

	// Step 3: Iterate through all logic operators to combine nodes.
	for i < len(nodes) {
		op := logics[i-1]

		// Case 1: Current operator is AND
		if op == rulepb.Logic_LOGIC_AND {
			// Build a chain of consecutive AND operations.
			andNode := nodes[i-1]
			for i < len(nodes) && logics[i-1] == rulepb.Logic_LOGIC_AND {
				// Combine the previous AND node with the next one.
				andNode = logic.NewLogic(andNode, logic.LogicAnd, nodes[i])
				i++
			}

			// After finishing a block of ANDs, check the previous operator type
			// to decide whether to attach this AND group to the current tree
			// with OR AND.

			// Potentially incorrect index; ensure this logic is valid.
			prevOp := logics[i-len(nodes)]

			if prevOp == rulepb.Logic_LOGIC_OR {
				current = logic.NewLogic(current, logic.LogicOr, andNode)
			} else {
				current = logic.NewLogic(current, logic.LogicAnd, andNode)
			}

		} else {
			// Case 2: Current operator is OR
			// Directly connect current node with the next one using OR logic.
			current = logic.NewLogic(current, logic.LogicOr, nodes[i])
			i++
		}
	}

	// Return the root of the constructed logical expression tree.
	return current
}

func newExpr(chain *rulepb.Expr) *exprs {
	return &exprs{
		chain: chain,
		tx: &tx{
			matched:   &rulepb.MatchedRules{},
			threshold: chain.Threshold,
		},
	}
}

func NewIngress() Rules {
	return &ingress{}
}

type ingress struct {
	expr sync.Map
}

func (in *ingress) Eval(ctx chttp.RequestContext, rule *rulepb.Rule) bool {
	zlog.Debugf("[edge][%s] >>> visit ingress eval", ctx.RequestId())

	if !rule.GetEnabled() {
		return false
	}

	cond := newCondition(rule.GetCondition())
	ruleCtx := newEvaluator(ctx)
	defer ruleCtx.Release()

	return cond.Accept(ruleCtx)
}

func (in *ingress) matched(ctx chttp.RequestContext,
	rule *rulepb.Rule) (id string, name string, score int32, ok bool) {

	if ok = in.Eval(ctx, rule); !ok {
		return "", "", 0, false
	}

	return rule.GetId(), rule.GetName(), rule.GetCondition().GetScore(), true
}

// EvalExpr a list of rule
func (in *ingress) EvalExpr(
	ctx chttp.RequestContext, chain *rulepb.Expr) (*rulepb.MatchedRules, bool) {
	zlog.Debugf("[edge][%s] >>> visit ingress", ctx.RequestId())

	if !chain.GetEnabled() {
		return nil, false
	}

	rules := chain.GetRules()
	if len(rules) == 1 {
		id, name, score, ok := in.matched(ctx, chain.GetRules()[0])
		if ok {
			return &rulepb.MatchedRules{
				Ids:    []string{id},
				Names:  []string{name},
				Scores: []int32{score},
			}, ok
		}

		return nil, ok
	}

	val, ok := in.expr.Load(chain.GetId())
	if !ok {
		val = newExpr(chain)
		in.expr.Store(chain.GetId(), val)
	}

	expr, _ := val.(*exprs)

	logicExpr := expr.build(in.matched)
	if ok = logicExpr.Eval(ctx); ok {
		return expr.tx.matched, ok
	}

	return nil, false
}
