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

	"github.com/jackc/pgx/v5"
	membermock "github.com/sentinez/controlplane/apps/tenant/v1/repos/members/mock"
	resourcemock "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources/mock"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/controlplane/pkg/headers/headerstest"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	coreRulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	_userId     = "user_1"
	_tenantId   = "tnt_1"
	_resourceId = "rsrc_1" // of _tenantId
	_foreignId  = "rsrc_2" // of another tenant
	_ruleId     = "rule_1"
	_foreignRul = "rule_2"
)

// rows keeps the rows of a repository by id, and records what is written.
type rows[M interface{ GetId() string }] struct {
	byId    map[string]M
	created []M
	deleted []string
}

func (r *rows[M]) Get(_ context.Context, id string) (M, error) {
	m, ok := r.byId[id]
	if !ok {
		return m, pgx.ErrNoRows
	}

	return m, nil
}

func (r *rows[M]) Create(_ context.Context, m M) (M, error) {
	r.created = append(r.created, m)
	return m, nil
}

func (r *rows[M]) Update(context.Context, M) error { return nil }

func (r *rows[M]) Delete(_ context.Context, id string) error {
	r.deleted = append(r.deleted, id)
	return nil
}

// The repos embed the wrappers of import_test.go, one level deeper than
// rows, so that the methods of rows win.
type authSecRules struct {
	secRuleRepo
	*rows[*securitypb.SecRule]
	listed []*securitypb.ListSecRulesRequest
}

func (f *authSecRules) List(_ context.Context,
	req *securitypb.ListSecRulesRequest,
) ([]*securitypb.SecRule, int64, error) {
	f.listed = append(f.listed, req)
	return nil, 0, nil
}

type authRateLimits struct {
	rateLimitRepo
	*rows[*securitypb.RateLimit]
}

type authCoreRulesets struct {
	coreRulesetRepo
	*rows[*securitypb.CoreRuleset]
}

// fixture is a service whose caller, _userId, has role in _tenantId and no
// other tenant. The rules _ruleId belong to _resourceId, and _foreignRul
// to _foreignId.
type fixture struct {
	svc    *SecurityService
	rules  *authSecRules
	limits *authRateLimits
	crs    *authCoreRulesets
}

func ownedBy(resourceId string) *typepb.Owner {
	return &typepb.Owner{ResourceId: resourceId, TenantId: _tenantId}
}

func newFixture(t *testing.T, role typepb.Role) *fixture {
	t.Helper()

	members := membermock.NewMockIMember(t)
	members.EXPECT().List(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, req *tenantpb.ListMemberRequest,
		) (*tenantpb.ListMemberResponse, error) {
			// like the repo: a filter of tenants keeps only those
			inTenant := len(req.GetTenantIds()) == 0 ||
				req.GetTenantIds()[0] == _tenantId
			if req.GetUserId() != _userId || !inTenant {
				return &tenantpb.ListMemberResponse{}, nil
			}

			return &tenantpb.ListMemberResponse{
				Members: []*tenantpb.Member{{
					UserId: _userId, TenantId: _tenantId,
					Roles: []typepb.Role{role},
				}},
			}, nil
		}).Maybe()

	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().Get(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, id string) (*tenantpb.Resource, error) {
			switch id {
			case _resourceId:
				return &tenantpb.Resource{Id: id, TenantId: _tenantId}, nil
			case _foreignId:
				return &tenantpb.Resource{Id: id, TenantId: "tnt_2"}, nil
			default:
				return nil, pgx.ErrNoRows
			}
		}).Maybe()
	resources.EXPECT().List(mock.Anything, mock.Anything).Return(
		&tenantpb.ListResourceResponse{
			Resources: []*tenantpb.Resource{{Id: _resourceId}},
		}, nil).Maybe()

	f := &fixture{
		rules: &authSecRules{rows: &rows[*securitypb.SecRule]{
			byId: map[string]*securitypb.SecRule{
				_ruleId:     {Id: _ruleId, Owner: ownedBy(_resourceId)},
				_foreignRul: {Id: _foreignRul, Owner: ownedBy(_foreignId)},
			}}},
		limits: &authRateLimits{rows: &rows[*securitypb.RateLimit]{
			byId: map[string]*securitypb.RateLimit{
				_ruleId:     {Id: _ruleId, Owner: ownedBy(_resourceId)},
				_foreignRul: {Id: _foreignRul, Owner: ownedBy(_foreignId)},
			}}},
		crs: &authCoreRulesets{rows: &rows[*securitypb.CoreRuleset]{
			byId: map[string]*securitypb.CoreRuleset{
				_ruleId:     {Id: _ruleId, Owner: ownedBy(_resourceId)},
				_foreignRul: {Id: _foreignRul, Owner: ownedBy(_foreignId)},
			}}},
	}
	f.svc = New(nil, f.rules, f.limits, f.crs, authz.New(members, resources))

	return f
}

func portalUser(t *testing.T) context.Context {
	t.Helper()

	return headerstest.User(t.Context(), _userId)
}

func adminUser(t *testing.T) context.Context {
	t.Helper()

	return headerstest.Admin(t.Context())
}

func secRuleOf(resourceId string) *securitypb.SecRule {
	return &securitypb.SecRule{Owner: &typepb.Owner{
		ResourceId: resourceId, TenantId: "tnt_evil",
	}}
}

func TestSecRuleAccess(t *testing.T) {
	const member = typepb.Role_ROLE_MEMBER
	const leader = typepb.Role_ROLE_LEADER

	tests := []struct {
		name     string
		giveRole typepb.Role
		call     func(*SecurityService, context.Context) error
		wantCode codes.Code
	}{
		{"member reads", member, func(s *SecurityService, c context.Context,
		) error {
			_, err := s.GetSecRule(c, &securitypb.GetSecRuleRequest{Id: _ruleId})
			return err
		}, codes.OK},
		{"member can not create", member, func(s *SecurityService,
			c context.Context) error {
			_, err := s.CreateSecRule(c, &securitypb.CreateSecRuleRequest{
				SecRule: secRuleOf(_resourceId)})
			return err
		}, codes.PermissionDenied},
		{"member can not update", member, func(s *SecurityService,
			c context.Context) error {
			_, err := s.UpdateSecRule(c, &securitypb.UpdateSecRuleRequest{
				Id: _ruleId, SecRule: &securitypb.SecRule{}})
			return err
		}, codes.PermissionDenied},
		{"member can not delete", member, func(s *SecurityService,
			c context.Context) error {
			_, err := s.DeleteSecRule(c,
				&securitypb.DeleteSecRuleRequest{Id: _ruleId})
			return err
		}, codes.PermissionDenied},
		{"leader creates", leader, func(s *SecurityService,
			c context.Context) error {
			_, err := s.CreateSecRule(c, &securitypb.CreateSecRuleRequest{
				SecRule: secRuleOf(_resourceId)})
			return err
		}, codes.OK},
		{"leader deletes", leader, func(s *SecurityService,
			c context.Context) error {
			_, err := s.DeleteSecRule(c,
				&securitypb.DeleteSecRuleRequest{Id: _ruleId})
			return err
		}, codes.OK},
		{"create on a foreign resource", leader, func(s *SecurityService,
			c context.Context) error {
			_, err := s.CreateSecRule(c, &securitypb.CreateSecRuleRequest{
				SecRule: secRuleOf(_foreignId)})
			return err
		}, codes.NotFound},
		{"get a foreign rule", leader, func(s *SecurityService,
			c context.Context) error {
			_, err := s.GetSecRule(c,
				&securitypb.GetSecRuleRequest{Id: _foreignRul})
			return err
		}, codes.NotFound},
		{"delete a foreign rule", leader, func(s *SecurityService,
			c context.Context) error {
			_, err := s.DeleteSecRule(c,
				&securitypb.DeleteSecRuleRequest{Id: _foreignRul})
			return err
		}, codes.NotFound},
		{"update a missing rule", leader, func(s *SecurityService,
			c context.Context) error {
			_, err := s.UpdateSecRule(c, &securitypb.UpdateSecRuleRequest{
				Id: "rule_x", SecRule: &securitypb.SecRule{}})
			return err
		}, codes.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, tt.giveRole)

			err := tt.call(f.svc, portalUser(t))

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestCreateSecRuleTakesTheTenantOfTheResource(t *testing.T) {
	f := newFixture(t, typepb.Role_ROLE_LEADER)

	_, err := f.svc.CreateSecRule(portalUser(t),
		&securitypb.CreateSecRuleRequest{SecRule: secRuleOf(_resourceId)})

	require.NoError(t, err)
	require.Len(t, f.rules.created, 1)
	assert.Equal(t, _tenantId, f.rules.created[0].GetOwner().GetTenantId())
}

func TestAdminControlPlaneTouchesAnyRule(t *testing.T) {
	f := newFixture(t, typepb.Role_ROLE_MEMBER)

	_, err := f.svc.DeleteSecRule(adminUser(t),
		&securitypb.DeleteSecRuleRequest{Id: _foreignRul})

	require.NoError(t, err)
	assert.Equal(t, []string{_foreignRul}, f.rules.deleted)
}

func TestListSecRulesIsScoped(t *testing.T) {
	f := newFixture(t, typepb.Role_ROLE_MEMBER)

	_, err := f.svc.ListSecRules(portalUser(t), &securitypb.ListSecRulesRequest{
		ResourceIds: []string{_resourceId, _foreignId},
	})

	require.NoError(t, err)
	require.Len(t, f.rules.listed, 1)
	assert.Equal(t, []string{_resourceId}, f.rules.listed[0].GetResourceIds())
}

func TestListSecRulesOfOthersRunsNoQuery(t *testing.T) {
	f := newFixture(t, typepb.Role_ROLE_MEMBER)

	_, err := f.svc.ListSecRules(portalUser(t), &securitypb.ListSecRulesRequest{
		ResourceIds: []string{_foreignId},
	})

	require.NoError(t, err)
	assert.Empty(t, f.rules.listed, "an empty filter would list every rule")
}

func TestListSecRulesWithoutFilterIsScoped(t *testing.T) {
	f := newFixture(t, typepb.Role_ROLE_MEMBER)

	_, err := f.svc.ListSecRules(portalUser(t), &securitypb.ListSecRulesRequest{})

	require.NoError(t, err)
	require.Len(t, f.rules.listed, 1)
	assert.Equal(t, []string{_resourceId}, f.rules.listed[0].GetResourceIds())
}

func TestListSecRulesAsAdminIsNotScoped(t *testing.T) {
	f := newFixture(t, typepb.Role_ROLE_MEMBER)

	_, err := f.svc.ListSecRules(adminUser(t), &securitypb.ListSecRulesRequest{})

	require.NoError(t, err)
	require.Len(t, f.rules.listed, 1)
	assert.Empty(t, f.rules.listed[0].GetResourceIds())
}

func TestRateLimitAndCoreRulesetAccess(t *testing.T) {
	validLimit := func(resourceId string) *securitypb.RateLimit {
		return &securitypb.RateLimit{
			TimeWindow: "10s", MaxRequests: 5, Owner: ownedBy(resourceId),
		}
	}
	validCrs := func(resourceId string) *securitypb.CoreRuleset {
		return &securitypb.CoreRuleset{
			Content: &coreRulesetpb.CoreRuleset{}, Owner: ownedBy(resourceId),
		}
	}

	tests := []struct {
		name     string
		giveRole typepb.Role
		call     func(*SecurityService, context.Context) error
		wantCode codes.Code
	}{
		{"get a foreign rate limit", typepb.Role_ROLE_ADMIN,
			func(s *SecurityService, c context.Context) error {
				_, err := s.GetRateLimit(c,
					&securitypb.GetRateLimitRequest{Id: _foreignRul})
				return err
			}, codes.NotFound},
		{"delete a foreign rate limit", typepb.Role_ROLE_ADMIN,
			func(s *SecurityService, c context.Context) error {
				_, err := s.DeleteRateLimit(c,
					&securitypb.DeleteRateLimitRequest{Id: _foreignRul})
				return err
			}, codes.NotFound},
		{"member can not create a rate limit", typepb.Role_ROLE_MEMBER,
			func(s *SecurityService, c context.Context) error {
				_, err := s.CreateRateLimit(c,
					&securitypb.CreateRateLimitRequest{
						RateLimit: validLimit(_resourceId)})
				return err
			}, codes.PermissionDenied},
		{"create a rate limit on a foreign resource", typepb.Role_ROLE_ADMIN,
			func(s *SecurityService, c context.Context) error {
				_, err := s.CreateRateLimit(c,
					&securitypb.CreateRateLimitRequest{
						RateLimit: validLimit(_foreignId)})
				return err
			}, codes.NotFound},
		{"get a foreign core ruleset", typepb.Role_ROLE_ADMIN,
			func(s *SecurityService, c context.Context) error {
				_, err := s.GetCoreRuleset(c,
					&securitypb.GetCoreRulesetRequest{Id: _foreignRul})
				return err
			}, codes.NotFound},
		{"member can not delete a core ruleset", typepb.Role_ROLE_MEMBER,
			func(s *SecurityService, c context.Context) error {
				_, err := s.DeleteCoreRuleset(c,
					&securitypb.DeleteCoreRulesetRequest{Id: _ruleId})
				return err
			}, codes.PermissionDenied},
		{"create a core ruleset on a foreign resource", typepb.Role_ROLE_ADMIN,
			func(s *SecurityService, c context.Context) error {
				_, err := s.CreateCoreRuleset(c,
					&securitypb.CreateCoreRulesetRequest{
						CoreRuleset: validCrs(_foreignId)})
				return err
			}, codes.NotFound},
		{"leader creates a core ruleset", typepb.Role_ROLE_LEADER,
			func(s *SecurityService, c context.Context) error {
				_, err := s.CreateCoreRuleset(c,
					&securitypb.CreateCoreRulesetRequest{
						CoreRuleset: validCrs(_resourceId)})
				return err
			}, codes.OK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t, tt.giveRole)

			err := tt.call(f.svc, portalUser(t))

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}
