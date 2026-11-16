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
	"time"

	cacertrepo "github.com/sentinez/controlplane/apps/tenant/v1/repos/cacertificates"
	memberrepo "github.com/sentinez/controlplane/apps/tenant/v1/repos/members"
	resourcerepo "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources"
	tenantrepo "github.com/sentinez/controlplane/apps/tenant/v1/repos/tenants"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/sznormalize"
	"github.com/sentinez/pkg/szprotobuf"
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
	cacert cacertrepo.ICaCertificate,
) *Service {
	return &Service{
		tx:       tx,
		resource: resource,
		member:   member,
		tenant:   tenant,
		cacert:   cacert,
		authz:    authz.New(member, resource),
		now:      time.Now,
	}
}

type Service struct {
	tx       *postgres.Tx
	resource resourcerepo.IResource
	member   memberrepo.IMember
	tenant   tenantrepo.ITenant
	cacert   cacertrepo.ICaCertificate
	authz    *authz.Authorizer
	// now is the clock the certificates are checked against
	now func() time.Time
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

// errNoDomain a resource stored without the domain the edge serves it at.
var errNoDomain = szerror.StatusInvalidArgumentF(
	"resource.domain is required")

// _resourceUpdatable the fields of a resource UpdateResource writes.
var _resourceUpdatable = []string{
	tenantpb.Resource_Name, tenantpb.Resource_Domain,
	tenantpb.Resource_Locations, tenantpb.Resource_Status,
}

// resourceOf is the tenant resource of the edge resource srv, its domain
// normalized as the edge matches it; errNoDomain without one.
func resourceOf(srv *edgepb.Resource) (*tenantpb.Resource, error) {
	rsc := sznormalize.TenantResource(srv)
	if rsc == nil {
		return nil, errNoDomain
	}

	if err := normalizeDomain(rsc); err != nil {
		return nil, err
	}

	return rsc, nil
}

// normalizeDomain sets the domain of rsc as the edge matches it,
// errNoDomain without one.
func normalizeDomain(rsc *tenantpb.Resource) error {
	rsc.Domain = sznormalize.ResourceDomain(rsc.GetDomain())
	if rsc.GetDomain() == "" {
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

	// the resource without its routes
	if req.GetDefault() {
		rsc.Locations = nil
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

	rsc, err := resourceOf(req.GetResource())
	if err != nil {
		return nil, err
	}

	rsc.Status = req.GetStatus()
	rsc.TenantId = req.GetTenantId()

	resp, err := svc.resource.Create(ctx, rsc)
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
	rsc, err := svc.authz.RequireResource(ctx, req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	paths, err := szprotobuf.ApplyMask(rsc, req.GetResource(),
		req.GetUpdateMask(), _resourceUpdatable...)
	if err != nil {
		return nil, err
	}

	if err = normalizeDomain(rsc); err != nil {
		return nil, err
	}

	err = svc.resource.Update(ctx, rsc, paths)
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
