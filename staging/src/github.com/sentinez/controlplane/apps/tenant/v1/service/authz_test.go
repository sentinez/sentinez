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

package tenantsvc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	membermock "github.com/sentinez/controlplane/apps/tenant/v1/repos/members/mock"
	resourcemock "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources/mock"
	tenantmock "github.com/sentinez/controlplane/apps/tenant/v1/repos/tenants/mock"
	"github.com/sentinez/controlplane/pkg/headers"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	_otherTenantId = "tnt_other"
	_resourceId    = "rsrc_1"
)

func roleOf(role typepb.Role) []typepb.Role { return []typepb.Role{role} }

// memberships makes the member repo answer the lookups of userId: it is a
// member of _tenantId with role, and of no other tenant.
func memberships(t *testing.T, userId string, role typepb.Role,
) *membermock.MockIMember {
	t.Helper()

	repo := membermock.NewMockIMember(t)
	repo.EXPECT().List(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, req *tenantpb.ListMemberRequest,
		) (*tenantpb.ListMemberResponse, error) {
			in := len(req.GetTenantIds()) == 0 ||
				req.GetTenantIds()[0] == _tenantId
			if !in || req.GetUserId() != userId {
				return &tenantpb.ListMemberResponse{}, nil
			}

			return &tenantpb.ListMemberResponse{
				Members: []*tenantpb.Member{{
					Id: "mem_self", UserId: userId, TenantId: _tenantId,
					Roles: roleOf(role),
				}},
			}, nil
		}).Maybe()

	return repo
}

func TestGetTenantOfOthers(t *testing.T) {
	svc := New(nil, nil, memberships(t, "user_1", typepb.Role_ROLE_ADMIN),
		tenantmock.NewMockITenant(t))

	_, err := svc.GetTenant(asUser(t, "user_1"),
		&tenantpb.GetTenantRequest{Id: _otherTenantId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestUpdateTenantNeedsAdmin(t *testing.T) {
	svc := New(nil, nil, memberships(t, "user_1", typepb.Role_ROLE_LEADER),
		tenantmock.NewMockITenant(t))

	_, err := svc.UpdateTenant(asUser(t, "user_1"),
		&tenantpb.UpdateTenantRequest{Id: _tenantId})

	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestDeleteTenantNeedsAdmin(t *testing.T) {
	svc := New(nil, nil, memberships(t, "user_1", typepb.Role_ROLE_MEMBER),
		tenantmock.NewMockITenant(t))

	_, err := svc.DeleteTenant(asUser(t, "user_1"),
		&tenantpb.DeleteTenantRequest{Id: _tenantId})

	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestListTenantIsScoped(t *testing.T) {
	tenants := tenantmock.NewMockITenant(t)
	tenants.EXPECT().List(mock.Anything, mock.MatchedBy(
		func(req *tenantpb.ListTenantRequest) bool {
			return len(req.GetIds()) == 1 && req.GetIds()[0] == _tenantId
		})).Return(&tenantpb.ListTenantResponse{}, nil)

	svc := New(nil, nil,
		memberships(t, "user_1", typepb.Role_ROLE_MEMBER), tenants)

	_, err := svc.ListTenant(asUser(t, "user_1"), &tenantpb.ListTenantRequest{})

	require.NoError(t, err)
}

func TestListTenantWithoutMembership(t *testing.T) {
	svc := New(nil, nil, memberships(t, "user_1", typepb.Role_ROLE_ADMIN),
		tenantmock.NewMockITenant(t)) // List must not run: no filter = all

	resp, err := svc.ListTenant(asUser(t, "user_9"),
		&tenantpb.ListTenantRequest{})

	require.NoError(t, err)
	assert.Empty(t, resp.GetTenants())
}

func TestListMemberIsScoped(t *testing.T) {
	members := membermock.NewMockIMember(t)
	members.EXPECT().List(mock.Anything, mock.MatchedBy(
		func(req *tenantpb.ListMemberRequest) bool {
			return req.GetUserId() == "user_1"
		})).RunAndReturn(
		func(_ context.Context, req *tenantpb.ListMemberRequest,
		) (*tenantpb.ListMemberResponse, error) {
			if len(req.GetTenantIds()) == 0 {
				return &tenantpb.ListMemberResponse{}, nil // own lookup
			}
			assert.Equal(t, []string{_tenantId}, req.GetTenantIds())

			return &tenantpb.ListMemberResponse{}, nil
		})

	_, err := New(nil, nil, members, nil).ListMember(asUser(t, "user_1"),
		&tenantpb.ListMemberRequest{UserId: "user_1"})

	require.NoError(t, err)
}

func TestCreateMemberTable(t *testing.T) {
	tests := []struct {
		name       string
		giveRole   typepb.Role
		giveTenant string
		giveGrant  typepb.Role
		wantCode   codes.Code
	}{
		{"leader can not add", typepb.Role_ROLE_LEADER, _tenantId,
			typepb.Role_ROLE_MEMBER, codes.PermissionDenied},
		{"admin of another tenant", typepb.Role_ROLE_ADMIN, _otherTenantId,
			typepb.Role_ROLE_MEMBER, codes.NotFound},
		{"admin can not give super admin", typepb.Role_ROLE_ADMIN, _tenantId,
			typepb.Role_ROLE_SUPER_ADMIN, codes.PermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(nil, nil, memberships(t, "user_1", tt.giveRole), nil)

			_, err := svc.CreateMember(asUser(t, "user_1"),
				&tenantpb.CreateMemberRequest{
					UserId: "user_2", TenantId: tt.giveTenant,
					Roles: roleOf(tt.giveGrant),
				})

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

// memberOf is a member the repo returns for Get, in a tenant with the
// given other members.
func memberOf(t *testing.T, target *tenantpb.Member, others ...*tenantpb.Member,
) *membermock.MockIMember {
	t.Helper()

	repo := membermock.NewMockIMember(t)
	repo.EXPECT().Get(mock.Anything, target.GetId()).Return(target, nil).Maybe()
	repo.EXPECT().List(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, req *tenantpb.ListMemberRequest,
		) (*tenantpb.ListMemberResponse, error) {
			all := append([]*tenantpb.Member{target}, others...)
			var found []*tenantpb.Member
			for _, m := range all {
				if req.GetUserId() == "" || req.GetUserId() == m.GetUserId() {
					found = append(found, m)
				}
			}

			return &tenantpb.ListMemberResponse{Members: found}, nil
		}).Maybe()
	repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Maybe()
	repo.EXPECT().Delete(mock.Anything, mock.Anything).Return(nil).Maybe()

	return repo
}

func adminMember(id, userId string) *tenantpb.Member {
	return &tenantpb.Member{
		Id: id, UserId: userId, TenantId: _tenantId,
		Roles: roleOf(typepb.Role_ROLE_ADMIN),
	}
}

func TestDeleteLastAdmin(t *testing.T) {
	self := adminMember("mem_1", "user_1")

	_, err := New(nil, nil, memberOf(t, self), nil).
		DeleteMember(asUser(t, "user_1"), &tenantpb.DeleteMemberRequest{
			Id: "mem_1",
		})

	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestDeleteAdminWithAnotherLeft(t *testing.T) {
	self := adminMember("mem_1", "user_1")
	other := adminMember("mem_2", "user_2")

	_, err := New(nil, nil, memberOf(t, other, self), nil).
		DeleteMember(asUser(t, "user_1"), &tenantpb.DeleteMemberRequest{
			Id: "mem_2",
		})

	require.NoError(t, err)
}

func TestDemoteLastAdmin(t *testing.T) {
	self := adminMember("mem_1", "user_1")

	_, err := New(nil, nil, memberOf(t, self), nil).
		UpdateMember(asUser(t, "user_1"), &tenantpb.UpdateMemberRequest{
			Id: "mem_1", Roles: roleOf(typepb.Role_ROLE_MEMBER),
		})

	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestUpdateMemberWithoutRolesKeepsLastAdmin(t *testing.T) {
	self := adminMember("mem_1", "user_1")

	_, err := New(nil, nil, memberOf(t, self), nil).
		UpdateMember(asUser(t, "user_1"), &tenantpb.UpdateMemberRequest{
			Id: "mem_1",
		})

	require.NoError(t, err)
}

func TestAdminCanNotTouchSuperAdmin(t *testing.T) {
	super := &tenantpb.Member{
		Id: "mem_2", UserId: "user_2", TenantId: _tenantId,
		Roles: roleOf(typepb.Role_ROLE_SUPER_ADMIN),
	}
	self := adminMember("mem_1", "user_1")

	_, err := New(nil, nil, memberOf(t, super, self), nil).
		DeleteMember(asUser(t, "user_1"), &tenantpb.DeleteMemberRequest{
			Id: "mem_2",
		})

	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestResourceNeedsRole(t *testing.T) {
	tests := []struct {
		name     string
		giveRole typepb.Role
		call     func(*Service) error
		wantCode codes.Code
	}{
		{"member can not update", typepb.Role_ROLE_MEMBER,
			func(s *Service) error {
				_, err := s.UpdateResource(asUser(t, "user_1"),
					&tenantpb.UpdateResourceRequest{Id: _resourceId})
				return err
			}, codes.PermissionDenied},
		{"member can not delete", typepb.Role_ROLE_MEMBER,
			func(s *Service) error {
				_, err := s.DeleteResource(asUser(t, "user_1"),
					&tenantpb.DeleteResourceRequest{Id: _resourceId})
				return err
			}, codes.PermissionDenied},
		{"member can not create", typepb.Role_ROLE_MEMBER,
			func(s *Service) error {
				_, err := s.CreateResource(asUser(t, "user_1"),
					&tenantpb.CreateResourceRequest{TenantId: _tenantId})
				return err
			}, codes.PermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := resourcemock.NewMockIResource(t)
			resources.EXPECT().Get(mock.Anything, _resourceId).
				Return(&tenantpb.Resource{
					Id: _resourceId, TenantId: _tenantId,
				}, nil).Maybe()
			svc := New(nil, resources, memberships(t, "user_1", tt.giveRole),
				nil)

			assert.Equal(t, tt.wantCode, status.Code(tt.call(svc)))
		})
	}
}

func TestResourceOfOthersIsNotFound(t *testing.T) {
	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().Get(mock.Anything, _resourceId).Return(
		&tenantpb.Resource{Id: _resourceId, TenantId: _otherTenantId}, nil)
	svc := New(nil, resources,
		memberships(t, "user_1", typepb.Role_ROLE_ADMIN), nil)

	_, err := svc.GetResource(asUser(t, "user_1"),
		&tenantpb.GetResourceRequest{Id: _resourceId})

	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.NotContains(t, err.Error(), _otherTenantId)
}

func TestResourceMissingIsNotFound(t *testing.T) {
	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().Get(mock.Anything, _resourceId).
		Return(nil, pgx.ErrNoRows)
	svc := New(nil, resources,
		memberships(t, "user_1", typepb.Role_ROLE_ADMIN), nil)

	_, err := svc.DeleteResource(asUser(t, "user_1"),
		&tenantpb.DeleteResourceRequest{Id: _resourceId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestListResourceIsScoped(t *testing.T) {
	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().List(mock.Anything, mock.MatchedBy(
		func(req *tenantpb.ListResourceRequest) bool {
			return len(req.GetTenantIds()) == 1 &&
				req.GetTenantIds()[0] == _tenantId
		})).Return(&tenantpb.ListResourceResponse{}, nil)
	svc := New(nil, resources,
		memberships(t, "user_1", typepb.Role_ROLE_MEMBER), nil)

	_, err := svc.ListResource(asUser(t, "user_1"),
		&tenantpb.ListResourceRequest{})

	require.NoError(t, err)
}

func TestListResourceOfOthers(t *testing.T) {
	svc := New(nil, resourcemock.NewMockIResource(t),
		memberships(t, "user_1", typepb.Role_ROLE_MEMBER), nil)

	resp, err := svc.ListResource(asUser(t, "user_1"),
		&tenantpb.ListResourceRequest{TenantIds: []string{_otherTenantId}})

	require.NoError(t, err)
	assert.Empty(t, resp.GetResources())
}

// tenantOnPlan is a repo whose tenant _tenantId is on plan; it expects an
// update only when wantUpdate.
func tenantOnPlan(t *testing.T, plan typepb.Plan, wantUpdate bool,
) *tenantmock.MockITenant {
	t.Helper()

	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().Get(mock.Anything, _tenantId).
		Return(&tenantpb.Tenant{Id: _tenantId, Plan: plan}, nil)

	if wantUpdate {
		repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
	}

	return repo
}

func TestUpdateTenantPlan(t *testing.T) {
	const free, std, pro = typepb.Plan_PLAN_FREE, typepb.Plan_PLAN_STANDARD,
		typepb.Plan_PLAN_PRO

	tests := []struct {
		name        string
		giveRole    typepb.Role
		giveCurrent typepb.Plan
		givePlan    typepb.Plan
		wantCode    codes.Code
	}{
		{"admin picks a higher plan", typepb.Role_ROLE_ADMIN, free, pro,
			codes.PermissionDenied},
		{"admin picks a lower plan", typepb.Role_ROLE_ADMIN, pro, std,
			codes.PermissionDenied},
		{"admin keeps the plan", typepb.Role_ROLE_ADMIN, std, std, codes.OK},
		{"admin leaves the plan out", typepb.Role_ROLE_ADMIN, std,
			typepb.Plan_PLAN_UNSPECIFIED, codes.OK},
		{"super admin changes it", typepb.Role_ROLE_SUPER_ADMIN, free, pro,
			codes.OK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(nil, nil, memberships(t, "user_1", tt.giveRole),
				tenantOnPlan(t, tt.giveCurrent, tt.wantCode == codes.OK))

			_, err := svc.UpdateTenant(asUser(t, "user_1"),
				&tenantpb.UpdateTenantRequest{
					Id: _tenantId, Name: "acme", Plan: tt.givePlan,
				})

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestUpdateTenantPlanAsAdminControlPlane(t *testing.T) {
	svc := New(nil, nil, nil, tenantOnPlan(t, typepb.Plan_PLAN_FREE, true))

	_, err := svc.UpdateTenant(asAdmin(t), &tenantpb.UpdateTenantRequest{
		Id: _tenantId, Plan: typepb.Plan_PLAN_PRO,
	})

	require.NoError(t, err)
}

func TestNewTenantPlan(t *testing.T) {
	const (
		unset = typepb.Plan_PLAN_UNSPECIFIED
		free  = typepb.Plan_PLAN_FREE
		pro   = typepb.Plan_PLAN_PRO

		portal = typepb.ControlPlane_CONTROL_PLANE_PORTAL
		admin  = typepb.ControlPlane_CONTROL_PLANE_ADMIN
	)

	tests := []struct {
		name      string
		givePlane typepb.ControlPlane
		givePlan  typepb.Plan
		wantPlan  typepb.Plan
		wantCode  codes.Code
	}{
		{"portal, none given", portal, unset, free, codes.OK},
		{"portal, free", portal, free, free, codes.OK},
		{"portal, a paid plan", portal, pro, unset, codes.PermissionDenied},
		{"admin control plane, a paid plan", admin, pro, pro, codes.OK},
		{"admin control plane, none given", admin, unset, free, codes.OK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &headers.Auth{Context: &typepb.Context{
				ControlPlane: tt.givePlane,
			}}

			got, err := newTenantPlan(auth, tt.givePlan)

			assert.Equal(t, tt.wantCode, status.Code(err))
			assert.Equal(t, tt.wantPlan, got)
		})
	}
}
