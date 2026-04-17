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
	"fmt"

	chttp "github.com/sentinez/core/http"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
)

type exprs struct {
	chain *rulepb.Expr
	root  *node
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
func (ex *exprs) build(exec MatchedFunc) (*node, error) {

	rules := ex.chain.GetRules()
	logics := ex.chain.GetLogics()

	if len(rules) == 0 {
		return nil, fmt.Errorf("expression contains no rules")
	}

	if len(logics)+1 != len(rules) {
		return nil, fmt.Errorf(
			"invalid expression: %d rules, %d logics",
			len(rules), len(logics))
	}

	nodes := make([]*node, len(rules))
	for i, r := range rules {
		nodes[i] = newNode(
			func(ctx chttp.RequestContext, out *rulepb.MatchedRules) bool {
				id, name, ok := exec(ctx, r)
				if ok && out != nil {
					out.Ids = append(out.Ids, id)
					out.Names = append(out.Names, name)
				}
				return ok
			})
	}

	var stack []*node
	stack = append(stack, nodes[0])

	for i := 1; i < len(nodes); i++ {
		op := logics[i-1]
		if op == rulepb.Logic_LOGIC_AND {
			left := stack[len(stack)-1]
			right := nodes[i]
			stack[len(stack)-1] = newLogic(left, logicAnd, right)
		} else {
			stack = append(stack, nodes[i])
		}
	}

	current := stack[0]
	for i := 1; i < len(stack); i++ {
		current = newLogic(current, logicOr, stack[i])
	}

	return current, nil
}

func newExpr(chain *rulepb.Expr) *exprs {
	return &exprs{
		chain: chain,
	}
}
