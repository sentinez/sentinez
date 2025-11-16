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
	corehttp "github.com/sentinez/sentinez/core/http"
	"github.com/sentinez/sentinez/core/internal/logic"
	"github.com/sentinez/sentinez/shared/zlog"
)

var _ Rules = (*ingress)(nil)

type Rules interface {
	Eval(ctx corehttp.RequestContext, rule *rulepb.Rule) bool
	EvalExpr(ctx corehttp.RequestContext, rule *rulepb.Expr) bool
}

type exprs struct {
	chain *rulepb.Expr
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
func (ex *exprs) build(
	exec func(corehttp.RequestContext, *rulepb.Rule) bool) *logic.Node {

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
		nodes[i] = logic.NewNode(func(ctx corehttp.RequestContext) bool {
			// zlog.Debugf("execute node %v", idx)
			return exec(ctx, r)
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
			// with OR or AND.

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

func NewIngress() Rules {
	return &ingress{}
}

type ingress struct {
	expr sync.Map
}

func (in *ingress) Eval(ctx corehttp.RequestContext, rule *rulepb.Rule) bool {
	zlog.Debugf("[edge][%s] >>> visit ingress eval", ctx.RequestId())

	if !rule.GetEnabled() {
		return false
	}

	cond := newCondition(rule.GetCondition())
	ruleCtx := newEvaluator(ctx)
	defer ruleCtx.Release()

	return cond.Accept(ruleCtx)
}

// EvalExpr a list of rule
func (in *ingress) EvalExpr(
	ctx corehttp.RequestContext, chain *rulepb.Expr) bool {

	zlog.Debugf("[edge][%s] >>> visit ingress", ctx.RequestId())

	if !chain.GetEnabled() {
		return false
	}

	rules := chain.GetRules()
	if len(rules) == 1 {
		return in.Eval(ctx, chain.GetRules()[0])
	}

	val, ok := in.expr.Load(chain.GetId())
	if !ok {
		val = &exprs{chain: chain}
		in.expr.Store(chain.GetId(), val)
	}

	expr, _ := val.(*exprs)

	return expr.build(in.Eval).Eval(ctx)
}
