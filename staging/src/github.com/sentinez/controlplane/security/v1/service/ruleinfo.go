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

package securitysvc

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"github.com/sentinez/core/modsec/catalog"
	"github.com/sentinez/core/storage/dbx/query"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/errorx"
)

// ── RuleInfo ───────────────────────────────────────────────────────────

// GetRuleInfo retrieves a rule of the OWASP CRS catalog by ID
func (srv *SecurityService) GetRuleInfo(_ context.Context,
	req *securitypb.GetRuleInfoRequest,
) (*securitypb.GetRuleInfoResponse, error) {
	rule, ok := catalog.Rule(req.GetVersion(), req.GetId())
	if !ok {
		return nil, errorx.StatusNotFoundF(
			"rule not found: id=%d", req.GetId())
	}

	return &securitypb.GetRuleInfoResponse{Rule: rule}, nil
}

// ListRuleInfos searches the OWASP CRS catalog, sorted by rule ID
func (srv *SecurityService) ListRuleInfos(_ context.Context,
	req *securitypb.ListRuleInfosRequest,
) (*securitypb.ListRuleInfosResponse, error) {
	match := ruleInfoMatcher(req)

	var rules []*corerulesetpb.RuleInfo
	for _, rule := range catalog.Rules(req.GetVersion()) {
		if match(rule) {
			rules = append(rules, rule)
		}
	}

	return &securitypb.ListRuleInfosResponse{
		Rules: pageOf(rules, req.GetPage()),
		Total: int64(len(rules)),
	}, nil
}

// ruleInfoMatcher reports whether a rule passes every filter of the
// request.
func ruleInfoMatcher(
	req *securitypb.ListRuleInfosRequest,
) func(*corerulesetpb.RuleInfo) bool {
	search := strings.ToLower(strings.TrimSpace(req.GetSearch()))

	severities := make([]string, 0, len(req.GetSeverities()))
	for _, s := range req.GetSeverities() {
		severities = append(severities, strings.ToUpper(s))
	}

	return func(rule *corerulesetpb.RuleInfo) bool {
		return (req.GetIncludeSystem() || !rule.GetSystem()) &&
			oneOf(req.GetIds(), rule.GetId()) &&
			oneOf(req.GetCategories(), rule.GetCategory()) &&
			oneOf(req.GetParanoiaLevels(), rule.GetParanoiaLevel()) &&
			oneOf(severities, rule.GetSeverity()) &&
			hasTags(rule, req.GetTags()) &&
			matchSearch(rule, search)
	}
}

// oneOf an empty filter matches everything.
func oneOf[T comparable](filter []T, value T) bool {
	return len(filter) == 0 || slices.Contains(filter, value)
}

func hasTags(rule *corerulesetpb.RuleInfo, tags []string) bool {
	for _, tag := range tags {
		if !slices.Contains(rule.GetTags(), tag) {
			return false
		}
	}

	return true
}

// matchSearch search is lower case: a part of the message or of a tag,
// or the leading digits of the ID.
func matchSearch(rule *corerulesetpb.RuleInfo, search string) bool {
	if search == "" {
		return true
	}

	id := strconv.FormatUint(uint64(rule.GetId()), 10)
	if strings.HasPrefix(id, search) ||
		strings.Contains(strings.ToLower(rule.GetMsg()), search) {
		return true
	}

	return slices.ContainsFunc(rule.GetTags(), func(tag string) bool {
		return strings.Contains(strings.ToLower(tag), search)
	})
}

// pageOf the page of the rules, everything without a page size.
func pageOf(
	rules []*corerulesetpb.RuleInfo, page *typepb.Pages,
) []*corerulesetpb.RuleInfo {
	size := int(page.GetSize())
	if size <= 0 {
		return rules
	}

	offset := query.GetOffset(int(page.GetIndex()), size)
	if offset >= len(rules) {
		return nil
	}

	return rules[offset:min(offset+size, len(rules))]
}
