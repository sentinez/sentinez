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
	"fmt"
	"slices"

	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/pkg/szerror"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

// ListMember implements tenantpb.TenantServiceServer.
func (svc *Service) ListMember(ctx context.Context,
	req *tenantpb.ListMemberRequest) (*tenantpb.ListMemberResponse, error) {

	req, ok, err := authz.ScopedTenants(ctx, svc.authz, req)
	if err != nil || !ok {
		return &tenantpb.ListMemberResponse{}, err
	}

	return svc.member.List(ctx, req)
}

// GetMember implements tenantpb.TenantServiceServer.
func (svc *Service) GetMember(ctx context.Context,
	req *tenantpb.GetMemberRequest) (*tenantpb.GetMemberResponse, error) {

	member, err := svc.getMember(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	err = svc.authz.RequireTenant(ctx, member.GetTenantId(), authz.Read)
	if err != nil {
		return nil, err
	}

	return &tenantpb.GetMemberResponse{Member: member}, nil
}

// CreateMember implements tenantpb.TenantServiceServer. A user can be a
// member of a tenant only once.
func (svc *Service) CreateMember(ctx context.Context,
	req *tenantpb.CreateMemberRequest,
) (*tenantpb.CreateMemberResponse, error) {

	err := svc.authz.RequireGrant(ctx, req.GetTenantId(), req.GetRoles())
	if err != nil {
		return nil, err
	}

	existing, err := svc.member.List(ctx, &tenantpb.ListMemberRequest{
		TenantIds: []string{req.GetTenantId()},
		UserId:    req.GetUserId(),
	})
	if err != nil {
		return nil, fmt.Errorf("member.List: %w", err)
	}

	if len(existing.GetMembers()) != 0 {
		return nil, szerror.StatusAlreadyExistsF(
			"member already exists: user_id=%s tenant_id=%s",
			req.GetUserId(), req.GetTenantId())
	}

	member, err := svc.member.Create(ctx, &tenantpb.Member{
		UserId:   req.GetUserId(),
		TenantId: req.GetTenantId(),
		Roles:    req.GetRoles(),
	})
	if err != nil {
		return nil, fmt.Errorf("member.Create: %w", err)
	}

	return &tenantpb.CreateMemberResponse{Member: member}, nil
}

// UpdateMember implements tenantpb.TenantServiceServer.
func (svc *Service) UpdateMember(ctx context.Context,
	req *tenantpb.UpdateMemberRequest,
) (*tenantpb.UpdateMemberResponse, error) {

	member, err := svc.getMember(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	// the caller must outrank the roles taken away as well as the new ones
	roles := append(slices.Clone(member.GetRoles()), req.GetRoles()...)
	err = svc.authz.RequireGrant(ctx, member.GetTenantId(), roles)
	if err != nil {
		return nil, err
	}

	// an update without roles leaves them as they are
	if len(req.GetRoles()) != 0 {
		err = svc.keepAdmin(ctx, member, req.GetRoles())
		if err != nil {
			return nil, err
		}
	}

	err = svc.member.Update(ctx, &tenantpb.Member{
		Id:    req.GetId(),
		Roles: req.GetRoles(),
	})
	if err != nil {
		return nil, fmt.Errorf("member.Update: %w", err)
	}

	return &tenantpb.UpdateMemberResponse{}, nil
}

// DeleteMember implements tenantpb.TenantServiceServer.
func (svc *Service) DeleteMember(ctx context.Context,
	req *tenantpb.DeleteMemberRequest,
) (*tenantpb.DeleteMemberResponse, error) {

	member, err := svc.getMember(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	err = svc.authz.RequireGrant(ctx, member.GetTenantId(), member.GetRoles())
	if err != nil {
		return nil, err
	}

	if err = svc.keepAdmin(ctx, member, nil); err != nil {
		return nil, err
	}

	if err = svc.member.Delete(ctx, req.GetId()); err != nil {
		return nil, fmt.Errorf("member.Delete: %w", err)
	}

	return &tenantpb.DeleteMemberResponse{}, nil
}

// getMember returns the member, or a NotFound status when it is missing.
func (svc *Service) getMember(
	ctx context.Context, id string) (*tenantpb.Member, error) {

	member, err := svc.member.Get(ctx, id)
	if szerror.IsNoRows(err) {
		return nil, szerror.StatusNotFoundF("member not found: id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("member.Get: %w", err)
	}

	return member, nil
}

// keepAdmin fails when member is the last admin of its tenant and would stop
// being one: with newRoles that rank lower, or with none (a removal). A
// tenant without an admin can not be managed any more.
func (svc *Service) keepAdmin(ctx context.Context, member *tenantpb.Member,
	newRoles []typepb.Role) error {

	manage := authz.Rank(authz.Manage)
	if authz.RankOf(member.GetRoles()) < manage {
		return nil
	}

	if authz.RankOf(newRoles) >= manage {
		return nil
	}

	resp, err := svc.member.List(ctx, &tenantpb.ListMemberRequest{
		TenantIds: []string{member.GetTenantId()},
	})
	if err != nil {
		return fmt.Errorf("member.List: %w", err)
	}

	for _, other := range resp.GetMembers() {
		if other.GetId() != member.GetId() &&
			authz.RankOf(other.GetRoles()) >= manage {

			return nil
		}
	}

	return szerror.StatusForbiddenF(
		"the last admin of tenant %s can not lose the role",
		member.GetTenantId())
}
