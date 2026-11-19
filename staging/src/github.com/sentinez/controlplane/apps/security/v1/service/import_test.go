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
	"slices"
	"testing"

	"github.com/sentinez/controlplane/apps/security/v1/repos/secratelimit"
	"github.com/sentinez/controlplane/apps/security/v1/repos/secrule"
	"github.com/sentinez/controlplane/apps/security/v1/repos/secruleset"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
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

func (f *fakeRepo[M]) Update(_ context.Context, m M, _ []string) error {
	f.updated = append(f.updated, m)
	return nil
}

// The fakes below adapt fakeRepo to the repository interfaces. Each embeds
// its interface one level deeper than fakeRepo, so that the methods of
// fakeRepo win; the others panic through the nil interface.

type (
	secRuleRepo      struct{ secrule.ISecRule }
	secRateLimitRepo struct{ secratelimit.ISecRateLimit }
	secRulesetRepo   struct{ secruleset.ISecRuleset }
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
	secRateLimitRepo
	*fakeRepo[*securitypb.SecRateLimit]
}

func (f fakeRateLimits) List(context.Context,
	*securitypb.ListSecRateLimitsRequest,
) ([]*securitypb.SecRateLimit, int64, error) {
	return f.fakeRepo.List()
}

type fakeRulesets struct {
	secRulesetRepo
	*fakeRepo[*securitypb.SecRuleset]
}

func (f fakeRulesets) List(context.Context,
	*securitypb.ListSecRulesetsRequest,
) ([]*securitypb.SecRuleset, int64, error) {
	return f.fakeRepo.List()
}

// newImportService returns a service over the fakes of rules, limits and
// crs.
func newImportService(rules *fakeRepo[*securitypb.SecRule],
	limits *fakeRepo[*securitypb.SecRateLimit],
	crs *fakeRepo[*securitypb.SecRuleset]) *SecurityService {
	return New(nil, fakeSecRules{fakeRepo: rules},
		fakeRateLimits{fakeRepo: limits}, fakeRulesets{fakeRepo: crs}, nil)
}

// importSecurity is a normalized proxy.yaml security block with a core
// ruleset, the sec rules "block" and "challenge" and the limiter "limit".
func importSecurity() *edgepb.Security {
	block := &rulepb.Action{Type: rulepb.ActionType_ACTION_TYPE_BLOCK}

	return &edgepb.Security{
		Ruleset: &rulesetpb.Setting{
			Crs: &rulesetpb.Crs{ParanoiaLevel: 1},
		},
		Secrules: []*edgepb.SecRule{
			{Ingress: &secrulepb.SecRule{Name: "block", Action: block}},
			{Ingress: &secrulepb.SecRule{Name: "challenge"}},
		},
		Limiters: []*edgepb.RateLimit{{
			Ingress:     &secrulepb.SecRule{Name: "limit"},
			TimeWindow:  "5s",
			MaxRequests: 100,
			Timeout:     "5s",
		}},
	}
}

func TestImportSecurityCreates(t *testing.T) {
	rules, limits := &fakeRepo[*securitypb.SecRule]{},
		&fakeRepo[*securitypb.SecRateLimit]{}
	crs := &fakeRepo[*securitypb.SecRuleset]{}

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
	limits := &fakeRepo[*securitypb.SecRateLimit]{
		stored: []*securitypb.SecRateLimit{{Id: "srlim_1", Name: "limit"}},
	}
	crs := &fakeRepo[*securitypb.SecRuleset]{stored: []*securitypb.SecRuleset{
		{Id: "srs_1"},
	}}

	err := newImportService(rules, limits, crs).
		ImportSecurity(t.Context(), _owner, "shop", importSecurity())

	require.NoError(t, err)
	require.Len(t, crs.updated, 1)
	assert.Equal(t, "srs_1", crs.updated[0].GetId())
	require.Len(t, rules.updated, 1)
	assert.Equal(t, "srule_1", rules.updated[0].GetId())
	require.Len(t, rules.created, 1)
	assert.Equal(t, "challenge", rules.created[0].GetName())
	require.Len(t, limits.updated, 1)
	assert.Equal(t, "srlim_1", limits.updated[0].GetId())
	assert.Empty(t, limits.created)
}

type importInvalidTest struct {
	name string
	give *edgepb.Security
}

func TestImportSecurityInvalid(t *testing.T) {
	tests := slices.Concat(importUnnamedTests(), importInvalidRuleTests())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newImportService(&fakeRepo[*securitypb.SecRule]{},
				&fakeRepo[*securitypb.SecRateLimit]{},
				&fakeRepo[*securitypb.SecRuleset]{})

			err := svc.ImportSecurity(t.Context(), _owner, "shop", tt.give)

			assert.Error(t, err)
		})
	}
}

// importUnnamedTests rules without a name: imports upsert by name.
func importUnnamedTests() []importInvalidTest {
	return []importInvalidTest{
		{"unnamed sec rule", &edgepb.Security{
			Secrules: []*edgepb.SecRule{
				{Ingress: &secrulepb.SecRule{}},
			},
		}},
		{"unnamed limiter", &edgepb.Security{
			Limiters: []*edgepb.RateLimit{{
				Ingress:    &secrulepb.SecRule{},
				TimeWindow: "5s", MaxRequests: 1,
			}},
		}},
	}
}

// importInvalidRuleTests named rules a create would reject.
func importInvalidRuleTests() []importInvalidTest {
	return []importInvalidTest{
		{"sec rule with an invalid expr", &edgepb.Security{
			Secrules: []*edgepb.SecRule{
				{Ingress: &secrulepb.SecRule{
					Name: "rule", Expr: invalidExpr(),
				}},
			},
		}},
		{"limiter with an invalid expr", &edgepb.Security{
			Limiters: []*edgepb.RateLimit{{
				Ingress: &secrulepb.SecRule{
					Name: "limit", Expr: invalidExpr(),
				},
				TimeWindow: "5s", MaxRequests: 1,
			}},
		}},
		{"limiter without max requests", &edgepb.Security{
			Limiters: []*edgepb.RateLimit{{
				Ingress:    &secrulepb.SecRule{Name: "limit"},
				TimeWindow: "5s",
			}},
		}},
	}
}
