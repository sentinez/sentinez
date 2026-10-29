// Copyright 2026 Sentinéz Labs.
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

package szrule

import (
	chttp "github.com/sentinez/core/szhttp"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
)

func buildEval(id string, expr *rulepb.Expression) EvalFunc {
	// expr is fixed once built: render it once, not on every match
	name := ExprString(expr)

	return func(ctx chttp.Request, mr *rulepb.MatchedRules) bool {
		if !execOrCond(ctx, expr.GetOrCondition()) {
			return false
		}

		if mr != nil {
			mr.Ids = append(mr.Ids, id)
			mr.Names = append(mr.Names, name)
		}

		return true
	}
}

func execOrCond(ctx chttp.Request, orCond []*rulepb.AndCondition) bool {
	if orCond == nil {
		return true
	}

	for _, andCond := range orCond {
		if execAndCond(ctx, andCond) &&
			execOrCond(ctx, andCond.GetOrCondition()) {
			return true
		}
	}

	return false
}

func execAndCond(ctx chttp.Request, andCond *rulepb.AndCondition) bool {
	if andCond == nil {
		return true
	}

	for _, rule := range andCond.GetRules() {
		if !eval(ctx, rule) {
			return false
		}
	}

	return true
}
