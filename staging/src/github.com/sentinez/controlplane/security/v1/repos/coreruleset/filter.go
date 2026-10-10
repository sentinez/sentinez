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

package coreruleset

import (
	"strings"

	sq "github.com/Masterminds/squirrel"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// _likeEscaper escapes the wildcards of a LIKE pattern.
var _likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// crs a partial setting, the content of a matching row contains it.
type crs = corerulesetpb.CoreRuleset

// listConditions maps the filters of the request to WHERE conditions,
// all of them must match.
func listConditions(
	req *securitypb.ListCoreRulesetsRequest,
) ([]sq.Sqlizer, error) {
	var conds []sq.Sqlizer

	if len(req.GetIds()) > 0 {
		conds = append(conds, sq.Eq{securitypb.CoreRuleset_Id: req.GetIds()})
	}

	if name := req.GetName(); name != "" {
		conds = append(conds, sq.ILike{
			securitypb.CoreRuleset_Name: "%" + _likeEscaper.Replace(name) + "%",
		})
	}

	for _, alternatives := range contentFilters(req) {
		cond, err := containsAny(alternatives)
		if err != nil {
			return nil, err
		}

		if cond != nil {
			conds = append(conds, cond)
		}
	}

	return conds, nil
}

// contentFilters the filters on the fields of content, each as the
// partial settings a row has to contain one of.
func contentFilters(req *securitypb.ListCoreRulesetsRequest) [][]*crs {
	filters := [][]*crs{
		partials(req.GetModes(), func(v corerulesetpb.EngineMode) *crs {
			return &crs{Mode: v}
		}),
		partials(req.GetVersions(), func(v corerulesetpb.Version) *crs {
			return &crs{Version: v}
		}),
		partials(req.GetStatuses(), statusPartial),
		partials(req.GetExclusionTags(), exclusionTagPartial),
		partials(req.GetExclusionRuleIds(), exclusionRulePartial),
		partials(req.GetOverrideRuleIds(), overrideRulePartial),
	}

	if level := req.GetParanoiaLevel(); level > 0 {
		filters = append(filters, []*crs{{ParanoiaLevel: level}})
	}

	// A single partial: the row has to list every category.
	if categories := req.GetCategories(); len(categories) > 0 {
		filters = append(filters, []*crs{{Categories: categories}})
	}

	return filters
}

func partials[T any](values []T, partial func(T) *crs) []*crs {
	out := make([]*crs, 0, len(values))
	for _, v := range values {
		out = append(out, partial(v))
	}

	return out
}

func statusPartial(v typepb.Status) *crs {
	return &crs{Status: v}
}

func exclusionTagPartial(tag string) *crs {
	return &crs{Exclusions: []*corerulesetpb.Exclusion{
		{Tags: []string{tag}},
	}}
}

func exclusionRulePartial(id uint32) *crs {
	return &crs{Exclusions: []*corerulesetpb.Exclusion{
		{Targets: []*corerulesetpb.ExclusionTarget{{RuleId: id}}},
	}}
}

func overrideRulePartial(id uint32) *crs {
	return &crs{Overrides: []*corerulesetpb.RuleOverride{{Id: id}}}
}

// containsAny the content contains one of the partial settings (JSONB
// containment), nil = no condition.
func containsAny(alternatives []*crs) (sq.Sqlizer, error) {
	if len(alternatives) == 0 {
		return nil, nil
	}

	or := make(sq.Or, 0, len(alternatives))

	for _, partial := range alternatives {
		data, err := _marshalContent.Marshal(partial)
		if err != nil {
			return nil, err
		}

		or = append(or, sq.Expr(
			securitypb.CoreRuleset_Content+" @> ?::jsonb", string(data),
		))
	}

	return or, nil
}
