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
	"bytes"
	"net/netip"
	"strings"

	chttp "github.com/sentinez/core/http"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
	"github.com/sentinez/shared/bytesconv"
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
		return matchRegex(des, bytesconv.S2b(src))
	default:
		return unmatched
	}
}

func matchBytes(op rulepb.Operator, src []byte, des string) bool {
	switch op {
	case rulepb.Operator_OPERATOR_EQ:
		return bytesconv.B2s(src) == des
	case rulepb.Operator_OPERATOR_NE:
		return bytesconv.B2s(src) != des
	case rulepb.Operator_OPERATOR_CONTAINS:
		return bytes.Contains(src, bytesconv.S2b(des))
	case rulepb.Operator_OPERATOR_PREFIX:
		return bytes.HasPrefix(src, bytesconv.S2b(des))
	case rulepb.Operator_OPERATOR_SUFFIX:
		return bytes.HasSuffix(src, bytesconv.S2b(des))
	case rulepb.Operator_OPERATOR_MATCHES:
		return matchRegex(des, src)
	default:
		return unmatched
	}
}

func matchSourcePath(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	des := cond.GetValue().GetStringValue()
	src := ctx.Path()

	return matchBytes(cond.GetOperator(), src, des)
}

func matchSourceBody(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	src := ctx.Body()
	des := cond.GetValue().GetStringValue()

	return matchBytes(cond.GetOperator(), src, des)
}

// nolint:funlen
func matchSourceHeader(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	key := cond.GetKey()
	val := cond.GetValue()
	op := cond.GetOperator()

	// If no key is provided, we check for existence of the listed keys
	if key == "" || key == "header" {
		list := val.GetListValue()
		if list == nil {
			// fallback to single key existence check
			s := val.GetStringValue()
			if s == "" {
				return unmatched
			}
			_, exist := ctx.Headers()[s]
			return matchOperator(op, exist)
		}

		des := list.GetValues()
		src := ctx.Headers()
		if len(src) == 0 {
			return unmatched
		}

		switch op {
		case rulepb.Operator_OPERATOR_IN:
			for _, v := range des {
				s := v.GetStringValue()
				if _, ok := src[s]; !ok {
					return unmatched
				}
			}
			return matched
		case rulepb.Operator_OPERATOR_NOT_IN:
			for _, v := range des {
				s := v.GetStringValue()
				if _, ok := src[s]; ok {
					return unmatched
				}
			}
			return matched
		}
		return unmatched
	}

	// If a key is provided, we check its value
	srcVal := ctx.Header(bytesconv.S2b(key))
	list := val.GetListValue()

	if list != nil {
		// Membership check
		found := false
		for _, v := range list.GetValues() {
			if bytes.Equal(srcVal, bytesconv.S2b(v.GetStringValue())) {
				found = true
				break
			}
		}
		return matchOperator(op, found)
	}

	// Single value check
	desString := val.GetStringValue()
	return matchBytes(op, srcVal, desString)
}

func matchSourceHost(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	src := ctx.Host()
	des := cond.GetValue().GetStringValue()

	return matchBytes(cond.GetOperator(), src, des)
}

// nolint:funlen
func matchSourceQuery(ctx chttp.RequestContext, cond *rulepb.Condition) bool {
	key := cond.GetKey()
	val := cond.GetValue()
	op := cond.GetOperator()

	// If no key is provided, we check for existence of the listed keys
	if key == "" || key == "query" {
		list := val.GetListValue()
		if list == nil {
			// fallback to single key existence check
			s := val.GetStringValue()
			if s == "" {
				return unmatched
			}
			_, exist := ctx.Queries()[s]
			return matchOperator(op, exist)
		}

		des := list.AsSlice()
		src := ctx.Queries()
		if len(src) == 0 {
			return unmatched
		}

		switch op {
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
		}
		return unmatched
	}

	// If a key is provided, we check its value
	srcVal := ctx.Query(bytesconv.S2b(key))
	list := val.GetListValue()

	if list != nil {
		// Membership check
		found := false
		for _, v := range list.GetValues() {
			if bytes.Equal(srcVal, bytesconv.S2b(v.GetStringValue())) {
				found = true
				break
			}
		}
		return matchOperator(op, found)
	}

	// Single value check
	desString := val.GetStringValue()
	return matchBytes(op, srcVal, desString)
}

func matchOperator(op rulepb.Operator, ok bool) bool {
	switch op {
	case rulepb.Operator_OPERATOR_IN:
		return ok
	case rulepb.Operator_OPERATOR_NOT_IN:
		return !ok
	case rulepb.Operator_OPERATOR_EQ:
		return ok
	case rulepb.Operator_OPERATOR_NE:
		return !ok
	default:
		return ok
	}
}

func matchIP(src []byte, des string) bool {
	prefix, err := netip.ParsePrefix(des)
	if err != nil {
		return bytesconv.B2s(src) == des
	}

	addr, err := netip.ParseAddr(bytesconv.B2s(src))
	if err != nil {
		return false
	}
	return prefix.Contains(addr)
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
				if bytes.Equal(src, bytesconv.S2b(v.GetStringValue())) {
					return matched
				}
			}
		}
		return unmatched
	case rulepb.Operator_OPERATOR_NOT_IN:
		list := cond.GetValue().GetListValue()
		if list != nil {
			for _, v := range list.GetValues() {
				if bytes.Equal(src, bytesconv.S2b(v.GetStringValue())) {
					return unmatched
				}
			}
		}
		return matched
	default:
		des := cond.GetValue().GetStringValue()
		return matchBytes(cond.GetOperator(), src, des)
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
