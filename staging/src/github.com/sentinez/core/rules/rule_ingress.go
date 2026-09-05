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
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
)

type MatchedFunc func(ctx chttp.RequestContext,
	rule *rulepb.Rule) (id string, name string, ok bool)

type EvalFunc func(ctx chttp.RequestContext, mr *rulepb.MatchedRules) bool

func NewEval(expr *rulepb.Expression) EvalFunc {
	return buildEval(expr, match)
}

func eval(ctx chttp.RequestContext, rule *rulepb.Rule) bool {
	return accept(ctx, rule.GetCondition())
}

func match(ctx chttp.RequestContext,
	rule *rulepb.Rule) (id string, name string, ok bool) {

	if ok = eval(ctx, rule); !ok {
		return "", "", false
	}

	return rule.GetId(), rule.GetName(), true
}
