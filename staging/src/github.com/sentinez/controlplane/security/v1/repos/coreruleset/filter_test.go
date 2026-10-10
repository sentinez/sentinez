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
	"testing"

	sq "github.com/Masterminds/squirrel"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type listReq = securitypb.ListCoreRulesetsRequest

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
			name:     "name",
			give:     &listReq{Name: "50%_off"},
			wantSQL:  "SELECT 1 WHERE name ILIKE ?",
			wantArgs: []any{`%50\%\_off%`},
		},
		{
			name: "modes",
			give: &listReq{Modes: []corerulesetpb.EngineMode{
				corerulesetpb.EngineMode_ENGINE_MODE_ON,
				corerulesetpb.EngineMode_ENGINE_MODE_DETECTION_ONLY,
			}},
			wantSQL: "SELECT 1 WHERE " +
				"(content @> ?::jsonb OR content @> ?::jsonb)",
			wantArgs: []any{
				`{"mode":"ENGINE_MODE_ON"}`,
				`{"mode":"ENGINE_MODE_DETECTION_ONLY"}`,
			},
		},
		{
			name: "version",
			give: &listReq{Versions: []corerulesetpb.Version{
				corerulesetpb.Version_VERSION_V4_17_0,
			}},
			wantSQL:  "SELECT 1 WHERE (content @> ?::jsonb)",
			wantArgs: []any{`{"version":"VERSION_V4_17_0"}`},
		},
		{
			name: "status",
			give: &listReq{Statuses: []typepb.Status{
				typepb.Status_STATUS_ACTIVE,
			}},
			wantSQL:  "SELECT 1 WHERE (content @> ?::jsonb)",
			wantArgs: []any{`{"status":"STATUS_ACTIVE"}`},
		},
		{
			name:     "paranoia level",
			give:     &listReq{ParanoiaLevel: 2},
			wantSQL:  "SELECT 1 WHERE (content @> ?::jsonb)",
			wantArgs: []any{`{"paranoia_level":2}`},
		},
		{
			name: "categories",
			give: &listReq{Categories: []corerulesetpb.Category{
				corerulesetpb.Category_CATEGORY_XSS,
				corerulesetpb.Category_CATEGORY_SQLI,
			}},
			wantSQL: "SELECT 1 WHERE (content @> ?::jsonb)",
			wantArgs: []any{
				`{"categories":["CATEGORY_XSS","CATEGORY_SQLI"]}`,
			},
		},
		{
			name:     "exclusion tag",
			give:     &listReq{ExclusionTags: []string{"attack-sqli"}},
			wantSQL:  "SELECT 1 WHERE (content @> ?::jsonb)",
			wantArgs: []any{`{"exclusions":[{"tags":["attack-sqli"]}]}`},
		},
		{
			name:    "exclusion rule",
			give:    &listReq{ExclusionRuleIds: []uint32{942100}},
			wantSQL: "SELECT 1 WHERE (content @> ?::jsonb)",
			wantArgs: []any{
				`{"exclusions":[{"targets":[{"rule_id":942100}]}]}`,
			},
		},
		{
			name:     "override rule",
			give:     &listReq{OverrideRuleIds: []uint32{920420}},
			wantSQL:  "SELECT 1 WHERE (content @> ?::jsonb)",
			wantArgs: []any{`{"overrides":[{"id":920420}]}`},
		},
		{
			name: "filters are combined",
			give: &listReq{
				Name:          "prod",
				ParanoiaLevel: 1,
				Modes: []corerulesetpb.EngineMode{
					corerulesetpb.EngineMode_ENGINE_MODE_ON,
				},
			},
			wantSQL: "SELECT 1 WHERE name ILIKE ? " +
				"AND (content @> ?::jsonb) AND (content @> ?::jsonb)",
			wantArgs: []any{
				"%prod%", `{"mode":"ENGINE_MODE_ON"}`, `{"paranoia_level":1}`,
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

func TestContentRoundTrip(t *testing.T) {
	give := &corerulesetpb.CoreRuleset{
		Mode:          corerulesetpb.EngineMode_ENGINE_MODE_ON,
		ParanoiaLevel: 2,
		Exclusions: []*corerulesetpb.Exclusion{
			{Tags: []string{"attack-sqli"}},
		},
	}

	data, err := marshalContent(give)
	require.NoError(t, err)

	got := &corerulesetpb.CoreRuleset{}
	require.NoError(t, _unmarshalContent.Unmarshal([]byte(data), got))
	assert.True(t, proto.Equal(give, got))

	empty, err := marshalContent(nil)
	require.NoError(t, err)
	assert.Equal(t, "{}", empty)
}
