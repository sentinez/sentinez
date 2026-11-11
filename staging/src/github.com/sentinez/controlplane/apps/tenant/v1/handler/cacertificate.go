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

	"github.com/sentinez/pkg/szlog"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
)

// ListCaCertificate implements tenantpb.TenantServiceServer.
func (t *Tenant) ListCaCertificate(ctx context.Context,
	req *tenantpb.ListCaCertificateRequest,
) (*tenantpb.ListCaCertificateResponse, error) {

	szlog.Debugf("[Tenant.ListCaCertificate] req = %v", req)

	return t.svc.ListCaCertificate(ctx, req)
}

// CreateCaCertificate implements tenantpb.TenantServiceServer. The request
// holds a private key: it is never logged.
func (t *Tenant) CreateCaCertificate(ctx context.Context,
	req *tenantpb.CreateCaCertificateRequest,
) (*tenantpb.CreateCaCertificateResponse, error) {

	szlog.Debugf("[Tenant.CreateCaCertificate] resource_id = %s",
		req.GetResourceId())

	return t.svc.CreateCaCertificate(ctx, req)
}

// GetCaCertificate implements tenantpb.TenantServiceServer.
func (t *Tenant) GetCaCertificate(ctx context.Context,
	req *tenantpb.GetCaCertificateRequest,
) (*tenantpb.GetCaCertificateResponse, error) {

	szlog.Debugf("[Tenant.GetCaCertificate] req = %v", req)

	return t.svc.GetCaCertificate(ctx, req)
}

// DeleteCaCertificate implements tenantpb.TenantServiceServer.
func (t *Tenant) DeleteCaCertificate(ctx context.Context,
	req *tenantpb.DeleteCaCertificateRequest,
) (*tenantpb.DeleteCaCertificateResponse, error) {

	szlog.Debugf("[Tenant.DeleteCaCertificate] req = %v", req)

	return t.svc.DeleteCaCertificate(ctx, req)
}
