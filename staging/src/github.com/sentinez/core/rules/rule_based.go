// Copyright 2025-2026 Duc-Hung Ho.
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

// nolint:funlen
func buildNode(pbNode *rulepb.RuleBased_Node, exec MatchedFunc) *node {
	if pbNode == nil {
		return nil
	}

	var nodes []*node

	// Build nodes for individual rules
	for _, r := range pbNode.GetRules() {
		nodes = append(nodes, newNode(
			func(ctx chttp.RequestContext, out *rulepb.MatchedRules) bool {
				id, name, ok := exec(ctx, r)
				if ok && out != nil {
					out.Ids = append(out.Ids, id)
					out.Names = append(out.Names, name)
				}
				return ok
			}),
		)
	}

	// Build nodes for sub-groups recursively
	for _, subNode := range pbNode.GetGroups() {
		subEvalNode := buildNode(subNode, exec)
		if subEvalNode != nil {
			nodes = append(nodes, subEvalNode)
		}
	}

	if len(nodes) == 0 {
		return nil
	}

	op := pbNode.GetOperator()

	// Separate handling for NOT
	if op == rulepb.Logic_LOGIC_NOT {
		// NOT usually applies to its children.
		// If multiple children, combine with AND then negate.
		current := nodes[0]
		for i := 1; i < len(nodes); i++ {
			current = newLogic(current, logicAnd, nodes[i])
		}
		return newLogic(current, logicNot, nil)
	}

	// For AND/OR, we chain them
	current := nodes[0]
	lOp := logicAnd
	if op == rulepb.Logic_LOGIC_OR {
		lOp = logicOr
	}

	for i := 1; i < len(nodes); i++ {
		current = newLogic(current, lOp, nodes[i])
	}

	return current
}
