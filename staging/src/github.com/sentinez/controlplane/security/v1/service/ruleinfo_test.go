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
	"testing"

	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type listRules = securitypb.ListRuleInfosRequest

func listRuleIDs(t *testing.T, req *listRules) ([]uint32, int64) {
	t.Helper()

	resp, err := New(nil, nil, nil, nil).
		ListRuleInfos(context.Background(), req)
	require.NoError(t, err)

	ids := make([]uint32, 0, len(resp.GetRules()))
	for _, rule := range resp.GetRules() {
		ids = append(ids, rule.GetId())
	}

	return ids, resp.GetTotal()
}

// nolint:funlen
func TestListRuleInfos(t *testing.T) {
	sqli := []corerulesetpb.Category{corerulesetpb.Category_CATEGORY_SQLI}

	tests := []struct {
		name      string
		give      *listRules
		wantIDs   []uint32
		wantTotal int64
	}{
		{
			name:      "ids",
			give:      &listRules{Ids: []uint32{942100, 941100, 1}},
			wantIDs:   []uint32{941100, 942100},
			wantTotal: 2,
		},
		{
			name:      "system rules are left out",
			give:      &listRules{Ids: []uint32{901001, 942100}},
			wantIDs:   []uint32{942100},
			wantTotal: 1,
		},
		{
			name: "system rules included",
			give: &listRules{
				Ids: []uint32{901001, 942100}, IncludeSystem: true,
			},
			wantIDs:   []uint32{901001, 942100},
			wantTotal: 2,
		},
		{
			name:      "search by message",
			give:      &listRules{Search: "via LIBINJECTION", Categories: sqli},
			wantIDs:   []uint32{942100, 942101},
			wantTotal: 2,
		},
		{
			name:      "search by id prefix",
			give:      &listRules{Search: "94210"},
			wantIDs:   []uint32{942100, 942101},
			wantTotal: 2,
		},
		{
			name: "filters are combined",
			give: &listRules{
				Categories:     sqli,
				ParanoiaLevels: []uint32{1},
				Severities:     []string{"critical"},
				Tags:           []string{"attack-sqli", "OWASP_CRS"},
				Search:         "9421",
				Page:           &typepb.Pages{Index: 1, Size: 2},
			},
			wantIDs:   []uint32{942100, 942140},
			wantTotal: 6,
		},
		{
			name: "second page",
			give: &listRules{
				Search: "94210", Page: &typepb.Pages{Index: 2, Size: 1},
			},
			wantIDs:   []uint32{942101},
			wantTotal: 2,
		},
		{
			name: "page out of range",
			give: &listRules{
				Search: "94210", Page: &typepb.Pages{Index: 9, Size: 5},
			},
			wantIDs:   []uint32{},
			wantTotal: 2,
		},
		{
			name: "version",
			give: &listRules{
				Version: corerulesetpb.Version_VERSION_V4_16_0,
				Categories: []corerulesetpb.Category{
					corerulesetpb.Category_CATEGORY_DATA_LEAKAGES_RUBY,
				},
			},
			wantIDs: []uint32{},
		},
		{
			name:    "tag not carried",
			give:    &listRules{Tags: []string{"attack-sqli", "attack-xss"}},
			wantIDs: []uint32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIDs, gotTotal := listRuleIDs(t, tt.give)

			assert.Equal(t, tt.wantIDs, gotIDs)
			assert.Equal(t, tt.wantTotal, gotTotal)
		})
	}
}

func TestGetRuleInfo(t *testing.T) {
	srv := New(nil, nil, nil, nil)

	resp, err := srv.GetRuleInfo(context.Background(),
		&securitypb.GetRuleInfoRequest{Id: 942100})
	require.NoError(t, err)
	assert.Equal(t, uint32(942100), resp.GetRule().GetId())
	assert.Equal(t,
		corerulesetpb.Category_CATEGORY_SQLI, resp.GetRule().GetCategory())
	assert.Contains(t, resp.GetRule().GetTags(), "attack-sqli")

	_, err = srv.GetRuleInfo(context.Background(),
		&securitypb.GetRuleInfoRequest{Id: 999999})
	assert.Equal(t, codes.NotFound, status.Code(err))
}
