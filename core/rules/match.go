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
	"net"

	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	"github.com/sentinez/sentinez/core/internal/zlog"
	"github.com/sentinez/sentinez/core/networks"
)

const (
	bypass    = false
	matched   = true
	unmatched = false
)

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
		return bypass
	}
}

func matchSourceQuery(ctx networks.Context, cond *ruleenginepb.Condition) bool {
	// debug, _ := protojson.Marshal(cond)
	// zlog.Debug(string(debug))

	des := cond.Value.GetListValue().AsSlice()
	src := ctx.Queries()
	zlog.Debugf("rules: src: %s -> des: %s", src, des)

	if len(src) == 0 {
		return bypass
	}

	switch cond.GetOperator() {
	case ruleenginepb.Operator_OPERATOR_IN:
		for _, d := range des {
			if ds, ok := d.(string); ok {
				if _, exist := src[ds]; !exist {
					return unmatched
				}
			}
		}

		return matched

	case ruleenginepb.Operator_OPERATOR_NOT_IN:
		for _, d := range des {
			if ds, ok := d.(string); ok {
				if _, exist := src[ds]; exist {
					return unmatched
				}
			}
		}

		return matched

	default:
		return bypass
	}
}

func matchSourceIP(ctx networks.Context, cond *ruleenginepb.Condition) bool {
	src := ctx.ClientIP()
	des := cond.GetValue().GetStringValue()

	switch cond.GetOperator() {
	case ruleenginepb.Operator_OPERATOR_EQ:
		_, ipnet, err := net.ParseCIDR(des)
		if err != nil {
			return src == des
		}

		return ipnet.Contains(net.ParseIP(src))
	case ruleenginepb.Operator_OPERATOR_NE:
		_, ipnet, err := net.ParseCIDR(des)
		if err != nil {
			return src != des
		}

		return !ipnet.Contains(net.ParseIP(src))
	default:
		return bypass
	}
}

func matchSourceMethod(ctx networks.Context, cond *ruleenginepb.Condition) bool {
	switch cond.GetOperator() {
	case ruleenginepb.Operator_OPERATOR_EQ:
		return bypass
	case ruleenginepb.Operator_OPERATOR_NE:
		return bypass
	default:
		return bypass
	}
}
