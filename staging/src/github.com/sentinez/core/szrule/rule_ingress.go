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

package szrule

import (
	chttp "github.com/sentinez/core/szhttp"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
)

type EvalFunc func(ctx chttp.Request, mr *rulepb.MatchedRules) bool

func NewEval(id string, expr *rulepb.Expression) EvalFunc {
	return buildEval(id, expr)
}

func eval(ctx chttp.Request, rule *rulepb.Rule) bool {
	return accept(ctx, rule.GetCondition())
}
