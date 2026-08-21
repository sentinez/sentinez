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

package corerule

import (
	chttp "github.com/sentinez/core/http"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
)

func buildEval(
	expr *rulepb.Expression,
	match MatchedFunc,
) func(ctx chttp.RequestContext, mr *rulepb.MatchedRules) bool {

	return func(ctx chttp.RequestContext, mr *rulepb.MatchedRules) bool {
		return execOrCond(expr.GetOrCondition(), match, ctx, mr)
	}
}

func execOrCond(
	orCond []*rulepb.AndCondition,
	match MatchedFunc,
	ctx chttp.RequestContext,
	mr *rulepb.MatchedRules,
) bool {
	if orCond == nil {
		return true
	}

	for _, andCond := range orCond {
		if execAndCond(andCond, match, ctx, mr) &&
			execOrCond(andCond.GetOrCondition(), match, ctx, mr) {
			return true
		}
	}

	return false
}

func execAndCond(
	andCond *rulepb.AndCondition,
	match MatchedFunc,
	ctx chttp.RequestContext,
	mr *rulepb.MatchedRules,
) bool {
	if andCond == nil {
		return true
	}

	var (
		isMatchAnd = true
		ids        []string
		names      []string
	)

	for _, rule := range andCond.GetRules() {
		id, name, isMatched := match(ctx, rule)
		if !isMatched {
			isMatchAnd = false
			break
		}

		ids = append(ids, id)
		names = append(names, name)
	}

	if isMatchAnd {
		if mr != nil {
			mr.Ids = append(mr.Ids, ids...)
			mr.Names = append(mr.Names, names...)
		}
		return true
	}

	return false
}
