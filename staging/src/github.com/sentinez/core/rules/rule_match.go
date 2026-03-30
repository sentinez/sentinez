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
	"net"
	"strings"

	chttp "github.com/sentinez/core/http"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
	"github.com/sentinez/shared/zlog"
)

const (
	bypass    = false
	matched   = true
	unmatched = false
)

func matchString(op rulepb.Operator, src, des string) bool {
	switch op {
	case rulepb.Operator_OPERATOR_EQ:
		return src == des
	case rulepb.Operator_OPERATOR_NE:
		return src != des
	case rulepb.Operator_OPERATOR_CONTAINS:
		return strings.Contains(src, des)
	case rulepb.Operator_OPERATOR_PREFIX:
		return strings.HasPrefix(src, des)
	case rulepb.Operator_OPERATOR_SUFFIX:
		return strings.HasSuffix(src, des)
	case rulepb.Operator_OPERATOR_MATCHES:
		return matchRegex(des, src)
	default:
		return unmatched
	}
}

func matchSourcePath(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	des := cond.GetValue().GetStringValue()
	src := ctx.Path()

	zlog.Debugf("rules: src: %s -> des: %s", src, des)
	return matchString(cond.GetOperator(), src, des)
}

func matchSourceBody(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	src := string(ctx.Body())
	des := cond.GetValue().GetStringValue()

	zlog.Debugf("rules body: srcLen: %d -> des: %s", len(src), des)
	return matchString(cond.GetOperator(), src, des)
}

func matchSourceHeader(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	list := cond.GetValue().GetListValue()
	if list != nil {
		src := ctx.Headers()
		switch cond.GetOperator() {
		case rulepb.Operator_OPERATOR_IN:
			for _, v := range list.GetValues() {
				s := v.GetStringValue()
				if _, ok := src[s]; !ok {
					return unmatched
				}
			}
			return matched

		case rulepb.Operator_OPERATOR_NOT_IN:
			for _, v := range list.GetValues() {
				s := v.GetStringValue()
				if _, ok := src[s]; ok {
					return unmatched
				}
			}
			return matched
		}
		return unmatched
	}

	key := cond.GetKey()
	if key != "" {
		val := ctx.Header(key)
		des := cond.GetValue().GetStringValue()
		return matchString(cond.GetOperator(), val, des)
	}

	return unmatched
}

func matchSourceHost(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	src := ctx.Host()
	des := cond.GetValue().GetStringValue()

	return matchString(cond.GetOperator(), src, des)
}

// nolint:funlen
func matchSourceQuery(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	list := cond.GetValue().GetListValue()
	if list != nil {
		des := list.AsSlice()
		src := ctx.Queries()

		if len(src) == 0 {
			return unmatched
		}

		switch cond.GetOperator() {
		case rulepb.Operator_OPERATOR_IN:
			for _, d := range des {
				if ds, ok := d.(string); ok {
					if _, exist := src[ds]; !exist {
						return unmatched
					}
				}
			}
			return matched

		case rulepb.Operator_OPERATOR_NOT_IN:
			for _, d := range des {
				if ds, ok := d.(string); ok {
					if _, exist := src[ds]; exist {
						return unmatched
					}
				}
			}
			return matched

		default:
			return unmatched
		}
	}

	key := cond.GetKey()
	if key != "" {
		val := ctx.Query(key)
		desString := cond.GetValue().GetStringValue()
		return matchString(cond.GetOperator(), val, desString)
	}

	return unmatched
}

func matchIP(src, des string) bool {
	_, ipnet, err := net.ParseCIDR(des)
	if err != nil {
		return src == des
	}

	return ipnet.Contains(net.ParseIP(src))
}

func matchSourceIP(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	src := ctx.RequestIP()

	switch cond.GetOperator() {
	case rulepb.Operator_OPERATOR_EQ:
		des := cond.GetValue().GetStringValue()
		return matchIP(src, des)
	case rulepb.Operator_OPERATOR_NE:
		des := cond.GetValue().GetStringValue()
		return !matchIP(src, des)
	case rulepb.Operator_OPERATOR_IN:
		list := cond.GetValue().GetListValue()
		if list != nil {
			for _, v := range list.GetValues() {
				if matchIP(src, v.GetStringValue()) {
					return matched
				}
			}
		}
		return unmatched
	case rulepb.Operator_OPERATOR_NOT_IN:
		list := cond.GetValue().GetListValue()
		if list != nil {
			for _, v := range list.GetValues() {
				if matchIP(src, v.GetStringValue()) {
					return unmatched
				}
			}
		}
		return matched
	default:
		return unmatched
	}
}

func matchSourceMethod(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	src := ctx.Method()

	switch cond.GetOperator() {
	case rulepb.Operator_OPERATOR_IN:
		list := cond.GetValue().GetListValue()
		if list != nil {
			for _, v := range list.GetValues() {
				if src == v.GetStringValue() {
					return matched
				}
			}
		}
		return unmatched
	case rulepb.Operator_OPERATOR_NOT_IN:
		list := cond.GetValue().GetListValue()
		if list != nil {
			for _, v := range list.GetValues() {
				if src == v.GetStringValue() {
					return unmatched
				}
			}
		}
		return matched
	default:
		des := cond.GetValue().GetStringValue()
		return matchString(cond.GetOperator(), src, des)
	}
}

func matchSourceTLS(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	des := cond.GetValue().GetStringValue()
	src := "false"
	if ctx.TLS() {
		src = "true"
	}

	switch cond.GetOperator() {
	case rulepb.Operator_OPERATOR_EQ:
		return src == des
	case rulepb.Operator_OPERATOR_NE:
		return src != des
	default:
		return unmatched
	}
}

func matchSourceJA4(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	src := ctx.JA4()
	if src == "" {
		return unmatched
	}
	des := cond.GetValue().GetStringValue()
	return matchString(cond.GetOperator(), src, des)
}
