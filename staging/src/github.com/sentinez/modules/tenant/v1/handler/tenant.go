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

// Package tenanthandler provides the Tenant service handler.
package tenanthandler

import (
	"context"

	tenantsvc "github.com/sentinez/modules/tenant/v1/service"
	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/tenant/v1"
)

var _ tenantpb.TenantServiceServer = (*Tenant)(nil)

func New(svc *tenantsvc.Service) tenantpb.TenantServiceServer {
	return &Tenant{
		svc: svc,
	}
}

// Tenant implement tenant.TenantServiceServer
type Tenant struct {
	svc *tenantsvc.Service
}

// GetResourceByDomain implements [tenantpb.TenantServiceServer].
func (t *Tenant) GetResourceByDomain(ctx context.Context,
	req *tenantpb.GetResourceByDomainRequest,
) (*tenantpb.GetResourceByDomainResponse, error) {
	return t.svc.GetResourceByDomain(ctx, req)
}

func (t *Tenant) GetResource(ctx context.Context,
	req *tenantpb.GetResourceRequest) (*tenantpb.GetResourceResponse, error) {
	return t.svc.GetResource(ctx, req)
}

// CreateResource implements tenantpb.TenantServiceServer.
func (t *Tenant) CreateResource(ctx context.Context,
	req *tenantpb.CreateResourceRequest,
) (*tenantpb.CreateResourceResponse, error) {

	return t.svc.CreateResource(ctx, req)
}

// DeleteResource implements tenantpb.TenantServiceServer.
func (t *Tenant) DeleteResource(ctx context.Context,
	req *tenantpb.DeleteResourceRequest,
) (*tenantpb.DeleteResourceResponse, error) {

	return t.svc.DeleteResource(ctx, req)
}

// UpdateResource implements tenantpb.TenantServiceServer.
func (t *Tenant) UpdateResource(ctx context.Context,
	req *tenantpb.UpdateResourceRequest,
) (*tenantpb.UpdateResourceResponse, error) {

	return t.svc.UpdateResource(ctx, req)
}

func (t *Tenant) ListResource(ctx context.Context,
	req *tenantpb.ListResourceRequest,
) (*tenantpb.ListResourceResponse, error) {

	return t.svc.ListResource(ctx, req)
}

// Status implement function of tenant.TenantServiceServer
func (t *Tenant) Status(ctx context.Context,
	req *tenantpb.StatusRequest,
) (*tenantpb.StatusResponse, error) {

	_, _ = ctx, req

	return &tenantpb.StatusResponse{Message: "OK"}, nil
}
