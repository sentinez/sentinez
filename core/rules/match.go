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

package rules

import (
	"strings"

	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	"github.com/sentinez/sentinez/core/networks"
)

const byPass = false

func matchSourcePath(ctx networks.Context, cond *ruleenginepb.Condition) bool {
	des := cond.GetValue().GetStringValue()
	src := ctx.Path()

	// zlog.Debugf("rules: src: %s -> des: %s", src, des)

	switch cond.GetOperator() {
	case ruleenginepb.Operator_OPERATOR_EQ:
		return src == des
	case ruleenginepb.Operator_OPERATOR_NE:
		return src != des
	default:
		return byPass
	}
}

func matchSourceQuery(ctx networks.Context, cond *ruleenginepb.Condition) bool {
	des := cond.GetValue().GetListValue().String()
	src := ctx.Queries()

	switch cond.GetOperator() {
	case ruleenginepb.Operator_OPERATOR_IN:
		for _, query := range src {
			if !strings.Contains(des, query) {
				return false
			}
		}
		return true
	case ruleenginepb.Operator_OPERATOR_NOT_IN:
		for _, query := range src {
			if strings.Contains(des, query) {
				return false
			}
		}
		return true
	default:
		return byPass
	}
}
