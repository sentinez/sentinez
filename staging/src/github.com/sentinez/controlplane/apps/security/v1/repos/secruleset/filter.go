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

package secruleset

import (
	"strings"

	sq "github.com/Masterminds/squirrel"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// _likeEscaper escapes the wildcards of a LIKE pattern.
var _likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// partial a partial setting, the setting of a matching row contains it.
type partial = rulesetpb.Setting

// crsPartial a partial setting on the CRS fields only.
func crsPartial(crs *rulesetpb.Crs) *partial {
	return &partial{Crs: crs}
}

// listConditions maps the filters of the request to WHERE conditions,
// all of them must match.
func listConditions(
	req *securitypb.ListSecRulesetsRequest,
) ([]sq.Sqlizer, error) {
	var conds []sq.Sqlizer

	if len(req.GetIds()) > 0 {
		conds = append(conds, sq.Eq{securitypb.SecRuleset_Id: req.GetIds()})
	}

	if ids := req.GetResourceIds(); len(ids) > 0 {
		conds = append(conds,
			sq.Eq{securitypb.SecRuleset_OwnerResourceId: ids})
	}

	if name := req.GetName(); name != "" {
		conds = append(conds, sq.ILike{
			securitypb.SecRuleset_Name: "%" + _likeEscaper.Replace(name) + "%",
		})
	}

	for _, alternatives := range rulesetFilters(req) {
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

// rulesetFilters the filters on the fields of the setting, each as the
// partial settings a row has to contain one of.
func rulesetFilters(
	req *securitypb.ListSecRulesetsRequest,
) [][]*partial {
	filters := [][]*partial{
		modePartials(req.GetModes()),
		partials(req.GetVersions(), func(v rulesetpb.CRSVersion) *partial {
			return crsPartial(&rulesetpb.Crs{Version: v})
		}),
		partials(req.GetStatuses(), statusPartial),
		partials(req.GetProviders(), providerPartial),
		partials(req.GetExclusionTags(), exclusionTagPartial),
		partials(req.GetExclusionRuleIds(), exclusionRulePartial),
		partials(req.GetOverrideRuleIds(), overrideRulePartial),
	}

	if level := req.GetParanoiaLevel(); level > 0 {
		filters = append(filters, []*partial{
			crsPartial(&rulesetpb.Crs{ParanoiaLevel: level}),
		})
	}

	// A single partial: the row has to list every category.
	if categories := req.GetCategories(); len(categories) > 0 {
		filters = append(filters, []*partial{
			crsPartial(&rulesetpb.Crs{Categories: categories}),
		})
	}

	return filters
}

func partials[T any](values []T, of func(T) *partial) []*partial {
	out := make([]*partial, 0, len(values))
	for _, v := range values {
		out = append(out, of(v))
	}

	return out
}

// modePartials a mode matches the engine mode of the CRS or of the
// managed ruleset.
func modePartials(modes []rulesetpb.EngineMode) []*partial {
	out := make([]*partial, 0, 2*len(modes))
	for _, v := range modes {
		out = append(out,
			crsPartial(&rulesetpb.Crs{Mode: v}),
			&partial{Managed: &rulesetpb.Managed{Mode: v}},
		)
	}

	return out
}

func providerPartial(v rulesetpb.Provider) *partial {
	return &partial{Provider: v}
}

func statusPartial(v typepb.Status) *partial {
	return &partial{Status: v}
}

func exclusionTagPartial(tag string) *partial {
	return crsPartial(&rulesetpb.Crs{Exclusions: []*rulesetpb.Exclusion{
		{Tags: []string{tag}},
	}})
}

func exclusionRulePartial(id uint32) *partial {
	return crsPartial(&rulesetpb.Crs{Exclusions: []*rulesetpb.Exclusion{
		{Targets: []*rulesetpb.ExclusionTarget{{RuleId: id}}},
	}})
}

func overrideRulePartial(id uint32) *partial {
	return crsPartial(&rulesetpb.Crs{
		Overrides: []*rulesetpb.RuleOverride{{Id: id}},
	})
}

// containsAny the setting contains one of the partial settings (JSONB
// containment), nil = no condition.
func containsAny(alternatives []*partial) (sq.Sqlizer, error) {
	if len(alternatives) == 0 {
		return nil, nil
	}

	or := make(sq.Or, 0, len(alternatives))

	for _, p := range alternatives {
		data, err := _marshalRuleset.Marshal(p)
		if err != nil {
			return nil, err
		}

		or = append(or, sq.Expr(
			securitypb.SecRuleset_Setting+" @> ?::jsonb", string(data),
		))
	}

	return or, nil
}
