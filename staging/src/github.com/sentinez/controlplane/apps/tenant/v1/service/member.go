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

	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/sentinez/shared/errorx"
)

// ListMember implements tenantpb.TenantServiceServer.
func (svc *Service) ListMember(ctx context.Context,
	req *tenantpb.ListMemberRequest) (*tenantpb.ListMemberResponse, error) {

	return svc.member.List(ctx, req)
}

// GetMember implements tenantpb.TenantServiceServer.
func (svc *Service) GetMember(ctx context.Context,
	req *tenantpb.GetMemberRequest) (*tenantpb.GetMemberResponse, error) {

	member, err := svc.getMember(ctx, req.GetId())
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

	existing, err := svc.member.List(ctx, &tenantpb.ListMemberRequest{
		TenantId: req.GetTenantId(),
		UserId:   req.GetUserId(),
	})
	if err != nil {
		return nil, fmt.Errorf("member.List: %w", err)
	}

	if len(existing.GetMembers()) != 0 {
		return nil, errorx.StatusAlreadyExistsF(
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

	if _, err := svc.getMember(ctx, req.GetId()); err != nil {
		return nil, err
	}

	err := svc.member.Update(ctx, &tenantpb.Member{
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

	if _, err := svc.getMember(ctx, req.GetId()); err != nil {
		return nil, err
	}

	if err := svc.member.Delete(ctx, req.GetId()); err != nil {
		return nil, fmt.Errorf("member.Delete: %w", err)
	}

	return &tenantpb.DeleteMemberResponse{}, nil
}

// getMember returns the member, or a NotFound status when it is missing.
func (svc *Service) getMember(
	ctx context.Context, id string) (*tenantpb.Member, error) {

	member, err := svc.member.Get(ctx, id)
	if errorx.IsNoRows(err) {
		return nil, errorx.StatusNotFoundF("member not found: id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("member.Get: %w", err)
	}

	return member, nil
}
