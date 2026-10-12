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

// ListTenant implements tenantpb.TenantServiceServer.
func (svc *Service) ListTenant(ctx context.Context,
	req *tenantpb.ListTenantRequest) (*tenantpb.ListTenantResponse, error) {

	return svc.tenant.List(ctx, req)
}

// GetTenant implements tenantpb.TenantServiceServer.
func (svc *Service) GetTenant(ctx context.Context,
	req *tenantpb.GetTenantRequest) (*tenantpb.GetTenantResponse, error) {

	tenant, err := svc.getTenant(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &tenantpb.GetTenantResponse{Tenant: tenant}, nil
}

// CreateTenant implements tenantpb.TenantServiceServer.
func (svc *Service) CreateTenant(ctx context.Context,
	req *tenantpb.CreateTenantRequest,
) (*tenantpb.CreateTenantResponse, error) {

	tenant, err := svc.tenant.Create(ctx, &tenantpb.Tenant{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Plan:        req.GetPlan(),
	})
	if err != nil {
		return nil, fmt.Errorf("tenant.Create: %w", err)
	}

	return &tenantpb.CreateTenantResponse{Tenant: tenant}, nil
}

// UpdateTenant implements tenantpb.TenantServiceServer.
func (svc *Service) UpdateTenant(ctx context.Context,
	req *tenantpb.UpdateTenantRequest,
) (*tenantpb.UpdateTenantResponse, error) {

	if _, err := svc.getTenant(ctx, req.GetId()); err != nil {
		return nil, err
	}

	err := svc.tenant.Update(ctx, &tenantpb.Tenant{
		Id:          req.GetId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Plan:        req.GetPlan(),
	})
	if err != nil {
		return nil, fmt.Errorf("tenant.Update: %w", err)
	}

	return &tenantpb.UpdateTenantResponse{}, nil
}

// DeleteTenant implements tenantpb.TenantServiceServer.
func (svc *Service) DeleteTenant(ctx context.Context,
	req *tenantpb.DeleteTenantRequest,
) (*tenantpb.DeleteTenantResponse, error) {

	if _, err := svc.getTenant(ctx, req.GetId()); err != nil {
		return nil, err
	}

	if err := svc.tenant.Delete(ctx, req.GetId()); err != nil {
		return nil, fmt.Errorf("tenant.Delete: %w", err)
	}

	return &tenantpb.DeleteTenantResponse{}, nil
}

// getTenant returns the tenant, or a NotFound status when it is missing.
func (svc *Service) getTenant(
	ctx context.Context, id string) (*tenantpb.Tenant, error) {

	tenant, err := svc.tenant.Get(ctx, id)
	if errorx.IsNoRows(err) {
		return nil, errorx.StatusNotFoundF("tenant not found: id=%s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("tenant.Get: %w", err)
	}

	return tenant, nil
}
