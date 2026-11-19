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
	"testing"

	sq "github.com/Masterminds/squirrel"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type listReq = securitypb.ListSecRulesetsRequest

// nolint:funlen
func TestListConditions(t *testing.T) {
	tests := []struct {
		name     string
		give     *listReq
		wantSQL  string
		wantArgs []any
	}{
		{name: "no filter", give: &listReq{}, wantSQL: "SELECT 1"},
		{
			name:     "ids",
			give:     &listReq{Ids: []string{"a", "b"}},
			wantSQL:  "SELECT 1 WHERE id IN (?,?)",
			wantArgs: []any{"a", "b"},
		},
		{
			name:     "resource ids",
			give:     &listReq{ResourceIds: []string{"a", "b"}},
			wantSQL:  "SELECT 1 WHERE owner_resource_id IN (?,?)",
			wantArgs: []any{"a", "b"},
		},
		{
			name:     "name",
			give:     &listReq{Name: "50%_off"},
			wantSQL:  "SELECT 1 WHERE name ILIKE ?",
			wantArgs: []any{`%50\%\_off%`},
		},
		{
			name: "modes",
			give: &listReq{Modes: []rulesetpb.EngineMode{
				rulesetpb.EngineMode_ENGINE_MODE_ON,
				rulesetpb.EngineMode_ENGINE_MODE_DETECTION_ONLY,
			}},
			wantSQL: "SELECT 1 WHERE (setting @> ?::jsonb OR " +
				"setting @> ?::jsonb OR setting @> ?::jsonb OR " +
				"setting @> ?::jsonb)",
			wantArgs: []any{
				`{"crs":{"mode":"ENGINE_MODE_ON"}}`,
				`{"managed":{"mode":"ENGINE_MODE_ON"}}`,
				`{"crs":{"mode":"ENGINE_MODE_DETECTION_ONLY"}}`,
				`{"managed":{"mode":"ENGINE_MODE_DETECTION_ONLY"}}`,
			},
		},
		{
			name: "providers",
			give: &listReq{Providers: []rulesetpb.Provider{
				rulesetpb.Provider_PROVIDER_MANAGED,
			}},
			wantSQL:  "SELECT 1 WHERE (setting @> ?::jsonb)",
			wantArgs: []any{`{"provider":"PROVIDER_MANAGED"}`},
		},
		{
			name: "version",
			give: &listReq{Versions: []rulesetpb.CRSVersion{
				rulesetpb.CRSVersion_CRS_VERSION_V4_17_0,
			}},
			wantSQL:  "SELECT 1 WHERE (setting @> ?::jsonb)",
			wantArgs: []any{`{"crs":{"version":"CRS_VERSION_V4_17_0"}}`},
		},
		{
			name: "status",
			give: &listReq{Statuses: []typepb.Status{
				typepb.Status_STATUS_ACTIVE,
			}},
			wantSQL:  "SELECT 1 WHERE (setting @> ?::jsonb)",
			wantArgs: []any{`{"status":"STATUS_ACTIVE"}`},
		},
		{
			name:     "paranoia level",
			give:     &listReq{ParanoiaLevel: 2},
			wantSQL:  "SELECT 1 WHERE (setting @> ?::jsonb)",
			wantArgs: []any{`{"crs":{"paranoia_level":2}}`},
		},
		{
			name: "categories",
			give: &listReq{Categories: []rulesetpb.Category{
				rulesetpb.Category_CATEGORY_XSS,
				rulesetpb.Category_CATEGORY_SQLI,
			}},
			wantSQL: "SELECT 1 WHERE (setting @> ?::jsonb)",
			wantArgs: []any{
				`{"crs":{"categories":["CATEGORY_XSS","CATEGORY_SQLI"]}}`,
			},
		},
		{
			name:    "exclusion tag",
			give:    &listReq{ExclusionTags: []string{"attack-sqli"}},
			wantSQL: "SELECT 1 WHERE (setting @> ?::jsonb)",
			wantArgs: []any{
				`{"crs":{"exclusions":[{"tags":["attack-sqli"]}]}}`,
			},
		},
		{
			name:    "exclusion rule",
			give:    &listReq{ExclusionRuleIds: []uint32{942100}},
			wantSQL: "SELECT 1 WHERE (setting @> ?::jsonb)",
			wantArgs: []any{
				`{"crs":{"exclusions":[{"targets":[{"rule_id":942100}]}]}}`,
			},
		},
		{
			name:     "override rule",
			give:     &listReq{OverrideRuleIds: []uint32{920420}},
			wantSQL:  "SELECT 1 WHERE (setting @> ?::jsonb)",
			wantArgs: []any{`{"crs":{"overrides":[{"id":920420}]}}`},
		},
		{
			name: "filters are combined",
			give: &listReq{
				Name:          "prod",
				ParanoiaLevel: 1,
				Modes: []rulesetpb.EngineMode{
					rulesetpb.EngineMode_ENGINE_MODE_ON,
				},
			},
			wantSQL: "SELECT 1 WHERE name ILIKE ? " +
				"AND (setting @> ?::jsonb OR setting @> ?::jsonb) " +
				"AND (setting @> ?::jsonb)",
			wantArgs: []any{
				"%prod%", `{"crs":{"mode":"ENGINE_MODE_ON"}}`,
				`{"managed":{"mode":"ENGINE_MODE_ON"}}`,
				`{"crs":{"paranoia_level":1}}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conds, err := listConditions(tt.give)
			require.NoError(t, err)

			builder := sq.Select("1")
			for _, cond := range conds {
				builder = builder.Where(cond)
			}

			gotSQL, gotArgs, err := builder.ToSql()
			require.NoError(t, err)
			assert.Equal(t, tt.wantSQL, gotSQL)
			assert.Equal(t, normalize(tt.wantArgs), normalize(gotArgs))
		})
	}
}

// normalize drops the insignificant spaces protojson adds at random.
func normalize(args []any) []any {
	out := make([]any, 0, len(args))
	for _, arg := range args {
		if s, ok := arg.(string); ok && strings.HasPrefix(s, "{") {
			arg = strings.ReplaceAll(s, " ", "")
		}

		out = append(out, arg)
	}

	return out
}

func TestRulesetRoundTrip(t *testing.T) {
	give := &rulesetpb.Setting{
		Provider: rulesetpb.Provider_PROVIDER_MANAGED,
		Crs: &rulesetpb.Crs{
			Mode:          rulesetpb.EngineMode_ENGINE_MODE_ON,
			ParanoiaLevel: 2,
			Exclusions: []*rulesetpb.Exclusion{
				{Tags: []string{"attack-sqli"}},
			},
		},
	}

	data, err := marshalRuleset(give)
	require.NoError(t, err)

	got := &rulesetpb.Setting{}
	require.NoError(t, protojson.Unmarshal([]byte(data), got))
	assert.True(t, proto.Equal(give, got))

	empty, err := marshalRuleset(nil)
	require.NoError(t, err)
	assert.Equal(t, "{}", empty)
}

func TestRulesetUnsetProviderIsOwasp(t *testing.T) {
	give := &rulesetpb.Setting{Status: typepb.Status_STATUS_ACTIVE}

	data, err := marshalRuleset(give)

	require.NoError(t, err)
	assert.JSONEq(t,
		`{"status":"STATUS_ACTIVE","provider":"PROVIDER_OWASP"}`, data)
	assert.Equal(t, rulesetpb.Provider_PROVIDER_UNSPECIFIED,
		give.GetProvider(), "the setting of the caller is not changed")
}
