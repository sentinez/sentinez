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
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sznormalize"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
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
		authz:    authz.New(member, resource),
	}
}

type Service struct {
	tx       *postgres.Tx
	resource resourcerepo.IResource
	member   memberrepo.IMember
	tenant   tenantrepo.ITenant
	authz    *authz.Authorizer
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

	found := rsrc.GetResources()[0]
	if err = svc.authz.RequireResourceOf(ctx, found, authz.Read); err != nil {
		return nil, err
	}

	return &tenantpb.GetResourceByDomainResponse{Resource: found}, nil
}

// errNoDomain a server stored without the domain the edge serves it at.
var errNoDomain = szerror.StatusInvalidArgumentF(
	"server.resource_domain is required")

// defaultServer is srv without its routes.
func defaultServer(srv *edgepb.Server) *edgepb.Server {
	return &edgepb.Server{
		Name:           srv.GetName(),
		ResourceId:     srv.GetResourceId(),
		ResourceDomain: srv.GetResourceDomain(),
	}
}

// normalizeServer normalizes the domain of srv as the edge matches it,
// errNoDomain without one.
func normalizeServer(srv *edgepb.Server) error {
	srv.ResourceDomain = sznormalize.ResourceDomain(srv.GetResourceDomain())
	if srv.GetResourceDomain() == "" {
		return errNoDomain
	}

	return nil
}

func (svc *Service) GetResource(ctx context.Context,
	req *tenantpb.GetResourceRequest) (*tenantpb.GetResourceResponse, error) {

	rsc, err := svc.authz.RequireResource(ctx, req.GetId(), authz.Read)
	if err != nil {
		return nil, err
	}

	if req.GetDefault() {
		rsc.Server = defaultServer(rsc.GetServer())
	}

	return &tenantpb.GetResourceResponse{Resource: rsc}, nil
}

// CreateResource implements tenantpb.TenantServiceServer.
func (svc *Service) CreateResource(ctx context.Context,
	req *tenantpb.CreateResourceRequest,
) (*tenantpb.CreateResourceResponse, error) {
	err := svc.authz.RequireTenant(ctx, req.GetTenantId(), authz.Write)
	if err != nil {
		return nil, err
	}

	if err = normalizeServer(req.GetServer()); err != nil {
		return nil, err
	}

	resp, err := svc.resource.Create(ctx, &tenantpb.Resource{
		Server:   req.GetServer(),
		Status:   req.GetStatus(),
		TenantId: req.GetTenantId(),
	})
	if err != nil {
		return nil, err
	}

	return &tenantpb.CreateResourceResponse{Resource: resp}, nil
}

// DeleteResource implements tenantpb.TenantServiceServer.
func (svc *Service) DeleteResource(ctx context.Context,
	req *tenantpb.DeleteResourceRequest,
) (*tenantpb.DeleteResourceResponse, error) {
	_, err := svc.authz.RequireResource(ctx, req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	if err = svc.resource.Delete(ctx, req.GetId()); err != nil {
		return nil, fmt.Errorf("resource.Delete: %w", err)
	}

	return &tenantpb.DeleteResourceResponse{}, nil
}

// UpdateResource implements tenantpb.TenantServiceServer.
func (svc *Service) UpdateResource(ctx context.Context,
	req *tenantpb.UpdateResourceRequest,
) (*tenantpb.UpdateResourceResponse, error) {
	_, err := svc.authz.RequireResource(ctx, req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	if srv := req.GetServer(); srv != nil {
		if err = normalizeServer(srv); err != nil {
			return nil, err
		}
	}

	err = svc.resource.Update(ctx, &tenantpb.Resource{
		Id:     req.GetId(),
		Status: req.GetStatus(),
		Server: req.GetServer(),
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

	req, ok, err := authz.ScopedTenants(ctx, svc.authz, req)
	if err != nil || !ok {
		return &tenantpb.ListResourceResponse{}, err
	}

	return svc.resource.List(ctx, req)
}
