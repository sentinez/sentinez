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

package tenanthandler

import (
	"context"

	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/sentinez/shared/zlog"
)

// ListTenant implements tenantpb.TenantServiceServer.
func (t *Tenant) ListTenant(ctx context.Context,
	req *tenantpb.ListTenantRequest,
) (*tenantpb.ListTenantResponse, error) {

	zlog.Debugf("[Tenant.ListTenant] req = %v", req)

	err := allow(ctx, tenantpb.GetTenantServiceListTenant())
	if err != nil {
		return nil, err
	}

	return t.svc.ListTenant(ctx, req)
}

// GetTenant implements tenantpb.TenantServiceServer.
func (t *Tenant) GetTenant(ctx context.Context,
	req *tenantpb.GetTenantRequest,
) (*tenantpb.GetTenantResponse, error) {

	zlog.Debugf("[Tenant.GetTenant] req = %v", req)

	err := allow(ctx, tenantpb.GetTenantServiceGetTenant())
	if err != nil {
		return nil, err
	}

	return t.svc.GetTenant(ctx, req)
}

// CreateTenant implements tenantpb.TenantServiceServer.
func (t *Tenant) CreateTenant(ctx context.Context,
	req *tenantpb.CreateTenantRequest,
) (*tenantpb.CreateTenantResponse, error) {

	zlog.Debugf("[Tenant.CreateTenant] req = %v", req)

	err := allow(ctx, tenantpb.GetTenantServiceCreateTenant())
	if err != nil {
		return nil, err
	}

	return t.svc.CreateTenant(ctx, req)
}

// UpdateTenant implements tenantpb.TenantServiceServer.
func (t *Tenant) UpdateTenant(ctx context.Context,
	req *tenantpb.UpdateTenantRequest,
) (*tenantpb.UpdateTenantResponse, error) {

	zlog.Debugf("[Tenant.UpdateTenant] req = %v", req)

	err := allow(ctx, tenantpb.GetTenantServiceUpdateTenant())
	if err != nil {
		return nil, err
	}

	return t.svc.UpdateTenant(ctx, req)
}

// DeleteTenant implements tenantpb.TenantServiceServer.
func (t *Tenant) DeleteTenant(ctx context.Context,
	req *tenantpb.DeleteTenantRequest,
) (*tenantpb.DeleteTenantResponse, error) {

	zlog.Debugf("[Tenant.DeleteTenant] req = %v", req)

	err := allow(ctx, tenantpb.GetTenantServiceDeleteTenant())
	if err != nil {
		return nil, err
	}

	return t.svc.DeleteTenant(ctx, req)
}
