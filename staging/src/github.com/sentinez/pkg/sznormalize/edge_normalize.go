// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sznormalize

import (
	"strconv"
	"strings"

	"github.com/sentinez/pkg/szprefix"
	"github.com/sentinez/pkg/szrand"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

const _defaultLimiterDuration = "5s"

// NormalizeEdgeSetting fills the defaults of edge in place: the ids its
// rules and conditions lack (proxy.yaml sets none), the durations of its
// limiters, the scalar values of its conditions as strings, and its
// resource domain as ResourceDomain gives it.
func NormalizeEdgeSetting(edge *edgepb.Setting) {
	if srv := edge.GetIngress().GetResource(); srv != nil {
		srv.Domain = ResourceDomain(srv.GetDomain())
	}

	for _, cdn := range edge.GetDelivery().GetCdn() {
		normalizeOrCond(cdn.GetRule().GetExpr().GetOrCondition())
	}

	normalizeEdgeSecurity(edge.GetSecurity())
}

// ResourceDomain is domain as the edge keys a setting by and matches the
// request host against: lower case, without the trailing dot.
func ResourceDomain(domain string) string {
	return strings.ToLower(strings.TrimSuffix(domain, "."))
}

func normalizeEdgeSecurity(edgeSec *edgepb.Security) {
	for _, rule := range edgeSec.GetSecrules() {
		normalizeSecRule(rule.GetIngress())
	}

	for _, lim := range edgeSec.GetLimiters() {
		normalizeLimiter(lim)
	}
}

func normalizeLimiter(lim *edgepb.RateLimit) {
	normalizeSecRule(lim.GetIngress())

	if lim.GetTimeWindow() == "" {
		lim.TimeWindow = _defaultLimiterDuration
	}

	if lim.GetTimeout() == "" {
		lim.Timeout = _defaultLimiterDuration
	}
}

// normalizeSecRule gives rule an id if it has none: a rule of proxy.yaml,
// one of the database keeps its own.
func normalizeSecRule(rule *secrulepb.SecRule) {
	if rule == nil {
		return
	}

	if rule.GetId() == "" {
		rule.Id = szrand.NewUUID(szprefix.SecRule)
	}

	normalizeOrCond(rule.GetExpr().GetOrCondition())
}

func normalizeOrCond(orCond []*rulepb.AndCondition) {
	for _, and := range orCond {
		for _, rule := range and.GetRules() {
			normalizeCondition(rule.GetCondition())
		}

		normalizeOrCond(and.GetOrCondition())
	}
}

func normalizeCondition(cond *rulepb.Condition) {
	if cond == nil {
		return
	}

	if cond.GetId() == "" {
		cond.Id = szrand.NewNanoId(szprefix.Condition)
	}

	cond.Value = stringValue(cond.GetValue())
}

// stringValue is val with its number and bool scalars, and those of its
// list, as strings: the rules match the request against strings only,
// while proxy.yaml reads `value: 443` as a number.
func stringValue(val *structpb.Value) *structpb.Value {
	switch kind := val.GetKind().(type) {
	case *structpb.Value_NumberValue:
		return structpb.NewStringValue(
			strconv.FormatFloat(kind.NumberValue, 'f', -1, 64))
	case *structpb.Value_BoolValue:
		return structpb.NewStringValue(
			strconv.FormatBool(kind.BoolValue))
	case *structpb.Value_ListValue:
		for i, v := range kind.ListValue.GetValues() {
			kind.ListValue.Values[i] = stringValue(v)
		}
	}

	return val
}
