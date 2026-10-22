// Copyright 2025 Sentinéz Labs.
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

	memberrepo "github.com/sentinez/controlplane/apps/tenant/v1/repos/members"
	resourcerepo "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources"
	tenantrepo "github.com/sentinez/controlplane/apps/tenant/v1/repos/tenants"
	"github.com/sentinez/core/szstorage/dbx/postgres"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/sentinez/shared/szerror"
	"github.com/sentinez/shared/szlog"
)

var _ tenantpb.TenantServiceServer = (*Service)(nil)

func New(
	tx *postgres.Tx,
	resource resourcerepo.IResource,
	member memberrepo.IMember,
	tenant tenantrepo.ITenant,
) *Service {
	return &Service{
		tx:       tx,
		resource: resource,
		member:   member,
		tenant:   tenant,
	}
}

type Service struct {
	tx       *postgres.Tx
	resource resourcerepo.IResource
	member   memberrepo.IMember
	tenant   tenantrepo.ITenant
}

// GetResourceByDomain implements [tenantpb.TenantServiceServer].
func (svc *Service) GetResourceByDomain(ctx context.Context,
	req *tenantpb.GetResourceByDomainRequest,
) (*tenantpb.GetResourceByDomainResponse, error) {

	rsrc, err := svc.resource.List(ctx, &tenantpb.ListResourceRequest{
		ResourceDomain: req.GetResourceDomain(),
	})
	if err != nil {
		return nil, err
	}

	if len(rsrc.GetResources()) == 0 {
		return nil, szerror.StatusNotFoundF("resource not found!")
	}

	resp := &tenantpb.GetResourceByDomainResponse{
		Resource: rsrc.GetResources()[0],
	}
	return resp, nil
}

func (svc *Service) defaultSetting() *tenantpb.ResourceSetting {
	st := &tenantpb.ResourceSetting{}
	return st
}

func (svc *Service) GetResource(ctx context.Context,
	req *tenantpb.GetResourceRequest) (*tenantpb.GetResourceResponse, error) {

	rsc, err := svc.resource.Get(ctx, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("resource.Get: %w", err)
	}

	if req.GetDefault() {
		rsc.ResourceSetting = svc.defaultSetting()
	}

	return &tenantpb.GetResourceResponse{Resource: rsc}, nil
}

// CreateResource implements tenantpb.TenantServiceServer.
func (svc *Service) CreateResource(ctx context.Context,
	req *tenantpb.CreateResourceRequest,
) (*tenantpb.CreateResourceResponse, error) {
	resource := &tenantpb.Resource{
		ResourceSetting: req.GetResourceSetting(),
		ResourceDomain:  req.GetResourceDomain(),
		ResourceName:    req.GetResourceName(),
		Status:          req.GetStatus(),
		TenantId:        req.GetTenantId(),
	}

	if resource.GetResourceSetting() == nil {
		resource.ResourceSetting = svc.defaultSetting()
	}

	resp, err := svc.resource.Create(ctx, resource)
	if err != nil {
		return nil, err
	}

	return &tenantpb.CreateResourceResponse{Resource: resp}, nil
}

// DeleteResource implements tenantpb.TenantServiceServer.
func (svc *Service) DeleteResource(ctx context.Context,
	req *tenantpb.DeleteResourceRequest,
) (*tenantpb.DeleteResourceResponse, error) {
	if err := svc.resource.Delete(ctx, req.GetId()); err != nil {
		return nil, fmt.Errorf("resource.Delete: %w", err)
	}

	return &tenantpb.DeleteResourceResponse{}, nil
}

// UpdateResource implements tenantpb.TenantServiceServer.
func (svc *Service) UpdateResource(ctx context.Context,
	req *tenantpb.UpdateResourceRequest,
) (*tenantpb.UpdateResourceResponse, error) {
	_, err := svc.resource.Get(ctx, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("resource.Get: %w", err)
	}

	err = svc.resource.Update(ctx, &tenantpb.Resource{
		Id:              req.GetId(),
		ResourceDomain:  req.GetResourceDomain(),
		ResourceName:    req.GetResourceName(),
		Status:          req.GetStatus(),
		ResourceSetting: req.GetResourceSetting(),
	})
	if err != nil {
		return nil, fmt.Errorf("resource.Update: %w", err)
	}

	return &tenantpb.UpdateResourceResponse{}, nil
}

func (svc *Service) Status(_ context.Context,
	_ *tenantpb.StatusRequest) (*tenantpb.StatusResponse, error) {
	return &tenantpb.StatusResponse{}, nil
}

func (svc *Service) ListResource(ctx context.Context,
	req *tenantpb.ListResourceRequest,
) (*tenantpb.ListResourceResponse, error) {
	szlog.Infof("tenant svc: req = %v", req)

	return svc.resource.List(ctx, req)
}
