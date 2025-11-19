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
)

type tx struct {
	matched *rulepb.MatchedRules
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
func (ex *exprs) build(exec MatchedFunc) *node {

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
	nodes := make([]*node, len(rules))
	for i, r := range rules {
		// idx := i // capture index for logging
		nodes[i] = newNode(func(ctx chttp.RequestContext) bool {
			id, name, ok := exec(ctx, r)
			if ok {
				ex.tx.matched.Ids = append(ex.tx.matched.Ids, id)
				ex.tx.matched.Names = append(ex.tx.matched.Names, name)
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
				andNode = newLogic(andNode, logicAnd, nodes[i])
				i++
			}

			// After finishing a block of ANDs, check the previous operator type
			// to decide whether to attach this AND group to the current tree
			// with OR AND.

			// Potentially incorrect index; ensure this logic is valid.
			prevOp := logics[i-len(nodes)]

			if prevOp == rulepb.Logic_LOGIC_OR {
				current = newLogic(current, logicOr, andNode)
			} else {
				current = newLogic(current, logicAnd, andNode)
			}

		} else {
			// Case 2: Current operator is OR
			// Directly connect current node with the next one using OR logic.
			current = newLogic(current, logicOr, nodes[i])
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
			matched: &rulepb.MatchedRules{},
		},
	}
}
