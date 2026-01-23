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

	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/tenant/v1"
	resourcerepo "github.com/sentinez/sentinez/internal/core/tenant/v1/repos/resources"
	"github.com/sentinez/shared/zlog"
)

var _ tenantpb.TenantServiceServer = (*Service)(nil)

func New(resource resourcerepo.IResource) *Service {
	return &Service{
		resource: resource,
	}
}

type Service struct {
	resource resourcerepo.IResource
}

func (svc *Service) Status(_ context.Context,
	_ *tenantpb.StatusRequest) (*tenantpb.StatusResponse, error) {
	return &tenantpb.StatusResponse{}, nil
}

func (svc *Service) ListResource(ctx context.Context,
	req *tenantpb.ListResourceRequest) (*tenantpb.ListResourceResponse, error) {
	zlog.Infof("tenant svc: req = %v", req)

	return svc.resource.List(ctx, req)
}
