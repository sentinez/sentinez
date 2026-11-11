// Copyright 2026 Duc-Hung Ho.
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

	"github.com/sentinez/controlplane/apps/security/v1/repos/coreruleset"
	"github.com/sentinez/controlplane/apps/security/v1/repos/ratelimit"
	"github.com/sentinez/controlplane/apps/security/v1/repos/secrule"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _owner = &typepb.Owner{TenantId: "tnt_a", ResourceId: "rsrc_a"}

// fakeRepo keeps the rows of a rule repository in memory.
type fakeRepo[M any] struct {
	stored  []M
	created []M
	updated []M
}

func (f *fakeRepo[M]) List() ([]M, int64, error) {
	return f.stored, int64(len(f.stored)), nil
}

func (f *fakeRepo[M]) Create(_ context.Context, m M) (M, error) {
	f.created = append(f.created, m)
	return m, nil
}

func (f *fakeRepo[M]) Update(_ context.Context, m M) error {
	f.updated = append(f.updated, m)
	return nil
}

// The fakes below adapt fakeRepo to the repository interfaces. Each embeds
// its interface one level deeper than fakeRepo, so that the methods of
// fakeRepo win; the others panic through the nil interface.

type (
	secRuleRepo     struct{ secrule.ISecRule }
	rateLimitRepo   struct{ ratelimit.IRateLimit }
	coreRulesetRepo struct{ coreruleset.ICoreRuleset }
)

type fakeSecRules struct {
	secRuleRepo
	*fakeRepo[*securitypb.SecRule]
}

func (f fakeSecRules) List(context.Context,
	*securitypb.ListSecRulesRequest) ([]*securitypb.SecRule, int64, error) {
	return f.fakeRepo.List()
}

type fakeRateLimits struct {
	rateLimitRepo
	*fakeRepo[*securitypb.RateLimit]
}

func (f fakeRateLimits) List(context.Context,
	*securitypb.ListRateLimitsRequest,
) ([]*securitypb.RateLimit, int64, error) {
	return f.fakeRepo.List()
}

type fakeCoreRulesets struct {
	coreRulesetRepo
	*fakeRepo[*securitypb.CoreRuleset]
}

func (f fakeCoreRulesets) List(context.Context,
	*securitypb.ListCoreRulesetsRequest,
) ([]*securitypb.CoreRuleset, int64, error) {
	return f.fakeRepo.List()
}

// newImportService returns a service over the fakes of rules, limits and
// crs.
func newImportService(rules *fakeRepo[*securitypb.SecRule],
	limits *fakeRepo[*securitypb.RateLimit],
	crs *fakeRepo[*securitypb.CoreRuleset]) *SecurityService {
	return New(nil, fakeSecRules{fakeRepo: rules},
		fakeRateLimits{fakeRepo: limits}, fakeCoreRulesets{fakeRepo: crs}, nil)
}

// importSecurity is a normalized proxy.yaml security block with a core
// ruleset, the sec rules "block" and "challenge" and the limiter "limit".
func importSecurity() *edgepb.Security {
	block := &rulepb.Action{Type: rulepb.ActionType_ACTION_TYPE_BLOCK}

	return &edgepb.Security{
		CoreRuleset: &corerulesetpb.CoreRuleset{ParanoiaLevel: 1},
		SecRules: []*edgepb.SecRule{
			{IngressRuntime: &secrulepb.SecRule{Name: "block", Action: block}},
			{IngressRuntime: &secrulepb.SecRule{Name: "challenge"}},
		},
		Limiters: []*edgepb.RateLimit{{
			IngressRuntime: &secrulepb.SecRule{Name: "limit"},
			TimeWindow:     "5s",
			MaxRequests:    100,
			Timeout:        "5s",
		}},
	}
}

func TestImportSecurityCreates(t *testing.T) {
	rules, limits := &fakeRepo[*securitypb.SecRule]{},
		&fakeRepo[*securitypb.RateLimit]{}
	crs := &fakeRepo[*securitypb.CoreRuleset]{}

	err := newImportService(rules, limits, crs).
		ImportSecurity(t.Context(), _owner, "shop", importSecurity())

	require.NoError(t, err)
	require.Len(t, crs.created, 1)
	assert.Equal(t, "shop", crs.created[0].GetName())
	assert.Equal(t, _owner, crs.created[0].GetOwner())
	require.Len(t, rules.created, 2)
	assert.Equal(t, "ACTION_TYPE_BLOCK", rules.created[0].GetAction())
	assert.Equal(t, _owner, rules.created[1].GetOwner())
	require.Len(t, limits.created, 1)
	assert.Equal(t, int64(100), limits.created[0].GetMaxRequests())
}

func TestImportSecurityUpdatesByName(t *testing.T) {
	rules := &fakeRepo[*securitypb.SecRule]{stored: []*securitypb.SecRule{
		{Id: "srule_1", Name: "block"},
	}}
	limits := &fakeRepo[*securitypb.RateLimit]{stored: []*securitypb.RateLimit{
		{Id: "rlim_1", Name: "limit"},
	}}
	crs := &fakeRepo[*securitypb.CoreRuleset]{stored: []*securitypb.CoreRuleset{
		{Id: "crs_1"},
	}}

	err := newImportService(rules, limits, crs).
		ImportSecurity(t.Context(), _owner, "shop", importSecurity())

	require.NoError(t, err)
	require.Len(t, crs.updated, 1)
	assert.Equal(t, "crs_1", crs.updated[0].GetId())
	require.Len(t, rules.updated, 1)
	assert.Equal(t, "srule_1", rules.updated[0].GetId())
	require.Len(t, rules.created, 1)
	assert.Equal(t, "challenge", rules.created[0].GetName())
	require.Len(t, limits.updated, 1)
	assert.Equal(t, "rlim_1", limits.updated[0].GetId())
	assert.Empty(t, limits.created)
}

func TestImportSecurityInvalid(t *testing.T) {
	tests := []struct {
		name string
		give *edgepb.Security
	}{
		{"unnamed sec rule", &edgepb.Security{
			SecRules: []*edgepb.SecRule{
				{IngressRuntime: &secrulepb.SecRule{}},
			},
		}},
		{"unnamed limiter", &edgepb.Security{
			Limiters: []*edgepb.RateLimit{{
				IngressRuntime: &secrulepb.SecRule{},
				TimeWindow:     "5s", MaxRequests: 1,
			}},
		}},
		{"limiter without max requests", &edgepb.Security{
			Limiters: []*edgepb.RateLimit{{
				IngressRuntime: &secrulepb.SecRule{Name: "limit"},
				TimeWindow:     "5s",
			}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newImportService(&fakeRepo[*securitypb.SecRule]{},
				&fakeRepo[*securitypb.RateLimit]{},
				&fakeRepo[*securitypb.CoreRuleset]{})

			err := svc.ImportSecurity(t.Context(), _owner, "shop", tt.give)

			assert.Error(t, err)
		})
	}
}
