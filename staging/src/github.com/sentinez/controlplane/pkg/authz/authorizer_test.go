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

package authz

import (
	"context"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/sentinez/controlplane/pkg/headers"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	_member = typepb.Role_ROLE_MEMBER
	_leader = typepb.Role_ROLE_LEADER
	_tadmin = typepb.Role_ROLE_ADMIN
	_super  = typepb.Role_ROLE_SUPER_ADMIN
)

// fakeMembers filters like the member repo: an empty filter matches all.
type fakeMembers struct {
	members []*tenantpb.Member
	calls   int
}

func (f *fakeMembers) List(_ context.Context, req *tenantpb.ListMemberRequest,
) (*tenantpb.ListMemberResponse, error) {
	f.calls++

	var found []*tenantpb.Member
	for _, m := range f.members {
		inTenants := len(req.GetTenantIds()) == 0 ||
			slices.Contains(req.GetTenantIds(), m.GetTenantId())
		isUser := req.GetUserId() == "" || req.GetUserId() == m.GetUserId()
		if inTenants && isUser {
			found = append(found, m)
		}
	}

	return &tenantpb.ListMemberResponse{Members: found}, nil
}

type fakeResources struct{ resources []*tenantpb.Resource }

func (f *fakeResources) Get(_ context.Context, id string,
) (*tenantpb.Resource, error) {
	for _, r := range f.resources {
		if r.GetId() == id {
			return r, nil
		}
	}

	return nil, pgx.ErrNoRows
}

func (f *fakeResources) List(_ context.Context,
	req *tenantpb.ListResourceRequest,
) (*tenantpb.ListResourceResponse, error) {
	var found []*tenantpb.Resource
	for _, r := range f.resources {
		if len(req.GetTenantIds()) == 0 ||
			slices.Contains(req.GetTenantIds(), r.GetTenantId()) {
			found = append(found, r)
		}
	}

	return &tenantpb.ListResourceResponse{Resources: found}, nil
}

func member(userId, tenantId string, role typepb.Role) *tenantpb.Member {
	return &tenantpb.Member{
		UserId: userId, TenantId: tenantId, Roles: []typepb.Role{role},
	}
}

func resource(id, tenantId string) *tenantpb.Resource {
	return &tenantpb.Resource{Id: id, TenantId: tenantId}
}

// newAuthorizer has user_1 as leader of tnt_a and member of tnt_b, user_2
// as admin of tnt_a, and user_3 in tnt_c.
func newAuthorizer() (*Authorizer, *fakeMembers) {
	members := &fakeMembers{members: []*tenantpb.Member{
		member("user_1", "tnt_a", _leader),
		member("user_1", "tnt_b", _member),
		member("user_2", "tnt_a", _tadmin),
		member("user_3", "tnt_c", _tadmin),
	}}
	resources := &fakeResources{resources: []*tenantpb.Resource{
		resource("rsrc_a1", "tnt_a"),
		resource("rsrc_a2", "tnt_a"),
		resource("rsrc_b1", "tnt_b"),
		resource("rsrc_c1", "tnt_c"),
	}}

	return New(members, resources), members
}

func asUser(userId string) context.Context {
	return headers.WithAuth(context.Background(), &headers.Auth{
		Context: &typepb.Context{UserId: userId, ControlPlane: _portal},
	})
}

func asAdmin() context.Context {
	return headers.WithAuth(context.Background(), &headers.Auth{
		Context: &typepb.Context{UserId: "root", ControlPlane: _admin},
	})
}

func TestRank(t *testing.T) {
	assert.Greater(t, Rank(_super), Rank(_tadmin))
	assert.Greater(t, Rank(_tadmin), Rank(_leader))
	assert.Greater(t, Rank(_leader), Rank(_member))
	assert.Greater(t, Rank(_member), Rank(typepb.Role_ROLE_UNSPECIFIED))
	assert.Equal(t, Rank(_leader), RankOf([]typepb.Role{_member, _leader}))
}

func TestRequireTenant(t *testing.T) {
	tests := []struct {
		name       string
		giveCtx    context.Context
		giveTenant string
		giveRole   typepb.Role
		wantCode   codes.Code
	}{
		{"member reads", asUser("user_1"), "tnt_b", Read, codes.OK},
		{"leader writes", asUser("user_1"), "tnt_a", Write, codes.OK},
		{"member can not write", asUser("user_1"), "tnt_b", Write,
			codes.PermissionDenied},
		{"leader can not manage", asUser("user_1"), "tnt_a", Manage,
			codes.PermissionDenied},
		{"outsider sees nothing", asUser("user_1"), "tnt_c", Read,
			codes.NotFound},
		{"admin control plane", asAdmin(), "tnt_c", Manage, codes.OK},
		{"no user in the token", asUser(""), "tnt_a", Read,
			codes.Unauthenticated},
		{"no caller", context.Background(), "tnt_a", Read,
			codes.Unauthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			az, _ := newAuthorizer()

			err := az.RequireTenant(tt.giveCtx, tt.giveTenant, tt.giveRole)

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestRequireTenant_AdminReadsNoMember(t *testing.T) {
	az, members := newAuthorizer()

	require.NoError(t, az.RequireTenant(asAdmin(), "tnt_a", Manage))
	assert.Zero(t, members.calls)
}

func TestRequireResource(t *testing.T) {
	tests := []struct {
		name         string
		giveCtx      context.Context
		giveResource string
		giveRole     typepb.Role
		wantCode     codes.Code
	}{
		{"leader writes", asUser("user_1"), "rsrc_a1", Write, codes.OK},
		{"member can not write", asUser("user_1"), "rsrc_b1", Write,
			codes.PermissionDenied},
		{"outsider", asUser("user_1"), "rsrc_c1", Read, codes.NotFound},
		{"missing", asUser("user_1"), "rsrc_x", Read, codes.NotFound},
		{"admin control plane", asAdmin(), "rsrc_c1", Write, codes.OK},
		{"admin control plane, missing", asAdmin(), "rsrc_x", Read,
			codes.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			az, _ := newAuthorizer()

			_, err := az.RequireResource(tt.giveCtx, tt.giveResource,
				tt.giveRole)

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestRequireResource_Returns(t *testing.T) {
	az, _ := newAuthorizer()

	got, err := az.RequireResource(asUser("user_1"), "rsrc_a2", Read)

	require.NoError(t, err)
	assert.Equal(t, "tnt_a", got.GetTenantId())
}

func TestRequireGrant(t *testing.T) {
	tests := []struct {
		name     string
		giveCtx  context.Context
		giveRole typepb.Role
		wantCode codes.Code
	}{
		{"admin gives leader", asUser("user_2"), _leader, codes.OK},
		{"admin gives admin", asUser("user_2"), _tadmin, codes.OK},
		{"admin can not give super admin", asUser("user_2"), _super,
			codes.PermissionDenied},
		{"leader can not manage", asUser("user_1"), _member,
			codes.PermissionDenied},
		{"admin control plane gives super admin", asAdmin(), _super,
			codes.OK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			az, _ := newAuthorizer()

			err := az.RequireGrant(tt.giveCtx, "tnt_a",
				[]typepb.Role{tt.giveRole})

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestTenantIds(t *testing.T) {
	tests := []struct {
		name          string
		giveCtx       context.Context
		giveRequested []string
		wantIds       []string
		wantNone      bool
	}{
		{"all of the caller", asUser("user_1"), nil,
			[]string{"tnt_a", "tnt_b"}, false},
		{"a part of them", asUser("user_1"), []string{"tnt_b"},
			[]string{"tnt_b"}, false},
		{"drops the foreign one", asUser("user_1"),
			[]string{"tnt_a", "tnt_c"}, []string{"tnt_a"}, false},
		{"only a foreign one", asUser("user_1"), []string{"tnt_c"}, nil, true},
		{"no membership", asUser("user_9"), nil, nil, true},
		{"admin control plane, unfiltered", asAdmin(), nil, nil, false},
		{"admin control plane, filtered", asAdmin(), []string{"tnt_c"},
			[]string{"tnt_c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			az, _ := newAuthorizer()

			got, err := az.TenantIds(tt.giveCtx, tt.giveRequested)

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantIds, got.Ids)
			assert.Equal(t, tt.wantNone, got.None)
		})
	}
}

func TestResourceIds(t *testing.T) {
	tests := []struct {
		name          string
		giveCtx       context.Context
		giveRequested []string
		wantIds       []string
		wantNone      bool
	}{
		{"all of the caller", asUser("user_1"), nil,
			[]string{"rsrc_a1", "rsrc_a2", "rsrc_b1"}, false},
		{"drops the foreign one", asUser("user_1"),
			[]string{"rsrc_a1", "rsrc_c1"}, []string{"rsrc_a1"}, false},
		{"only a foreign one", asUser("user_1"), []string{"rsrc_c1"}, nil,
			true},
		{"no membership", asUser("user_9"), nil, nil, true},
		{"admin control plane", asAdmin(), nil, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			az, _ := newAuthorizer()

			got, err := az.ResourceIds(tt.giveCtx, tt.giveRequested)

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantIds, got.Ids)
			assert.Equal(t, tt.wantNone, got.None)
		})
	}
}

func TestScopes_NoUserInToken(t *testing.T) {
	az, _ := newAuthorizer()

	_, tenantErr := az.TenantIds(asUser(""), nil)
	_, resourceErr := az.ResourceIds(asUser(""), nil)

	assert.Equal(t, codes.Unauthenticated, status.Code(tenantErr))
	assert.Equal(t, codes.Unauthenticated, status.Code(resourceErr))
}

func TestCoMemberUserIds(t *testing.T) {
	az, _ := newAuthorizer()

	got, err := az.CoMemberUserIds(asUser("user_1"))

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"user_1", "user_2"}, got)
}

func TestCoMemberUserIds_NoMembership(t *testing.T) {
	az, _ := newAuthorizer()

	got, err := az.CoMemberUserIds(asUser("user_9"))

	require.NoError(t, err)
	assert.Equal(t, []string{"user_9"}, got)
}

func TestScoped(t *testing.T) {
	tests := []struct {
		name          string
		giveCtx       context.Context
		giveRequested []string
		wantIds       []string
	}{
		{"fills in all of the caller", asUser("user_1"), nil,
			[]string{"rsrc_a1", "rsrc_a2", "rsrc_b1"}},
		{"drops the foreign one", asUser("user_1"),
			[]string{"rsrc_a1", "rsrc_c1"}, []string{"rsrc_a1"}},
		{"admin control plane keeps it", asAdmin(), []string{"rsrc_c1"},
			[]string{"rsrc_c1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			az, _ := newAuthorizer()
			give := &securitypb.ListSecRulesRequest{
				ResourceIds: tt.giveRequested,
			}

			got, ok, err := Scoped(tt.giveCtx, az, give)

			require.NoError(t, err)
			assert.True(t, ok)
			assert.ElementsMatch(t, tt.wantIds, got.GetResourceIds())
		})
	}
}

func TestScoped_DoesNotChangeTheRequest(t *testing.T) {
	az, _ := newAuthorizer()
	give := &securitypb.ListSecRulesRequest{
		ResourceIds: []string{"rsrc_a1", "rsrc_c1"},
	}

	_, _, err := Scoped(asUser("user_1"), az, give)

	require.NoError(t, err)
	assert.Equal(t, []string{"rsrc_a1", "rsrc_c1"}, give.GetResourceIds())
}

func TestScoped_NothingToSee(t *testing.T) {
	az, _ := newAuthorizer()
	give := &securitypb.ListSecRulesRequest{ResourceIds: []string{"rsrc_c1"}}

	_, ok, err := Scoped(asUser("user_1"), az, give)

	require.NoError(t, err)
	assert.False(t, ok)
}
