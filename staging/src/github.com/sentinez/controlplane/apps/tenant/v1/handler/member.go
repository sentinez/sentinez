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

	"github.com/sentinez/controlplane/pkg/headers"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/zlog"
)

// ListMember implements tenantpb.TenantServiceServer.
func (t *Tenant) ListMember(ctx context.Context,
	req *tenantpb.ListMemberRequest) (*tenantpb.ListMemberResponse, error) {

	zlog.Debugf("[Tenant.ListMember] req = %v", req)

	if err := allow(ctx, tenantpb.GetTenantServiceListMember()); err != nil {
		return nil, err
	}

	return t.svc.ListMember(ctx, req)
}

// GetMember implements tenantpb.TenantServiceServer.
func (t *Tenant) GetMember(ctx context.Context,
	req *tenantpb.GetMemberRequest) (*tenantpb.GetMemberResponse, error) {

	zlog.Debugf("[Tenant.GetMember] req = %v", req)

	if err := allow(ctx, tenantpb.GetTenantServiceGetMember()); err != nil {
		return nil, err
	}

	return t.svc.GetMember(ctx, req)
}

// CreateMember implements tenantpb.TenantServiceServer.
func (t *Tenant) CreateMember(ctx context.Context,
	req *tenantpb.CreateMemberRequest,
) (*tenantpb.CreateMemberResponse, error) {

	zlog.Debugf("[Tenant.CreateMember] req = %v", req)

	err := allow(ctx, tenantpb.GetTenantServiceCreateMember())
	if err != nil {
		return nil, err
	}

	return t.svc.CreateMember(ctx, req)
}

// UpdateMember implements tenantpb.TenantServiceServer.
func (t *Tenant) UpdateMember(ctx context.Context,
	req *tenantpb.UpdateMemberRequest,
) (*tenantpb.UpdateMemberResponse, error) {

	zlog.Debugf("[Tenant.UpdateMember] req = %v", req)

	err := allow(ctx, tenantpb.GetTenantServiceUpdateMember())
	if err != nil {
		return nil, err
	}

	return t.svc.UpdateMember(ctx, req)
}

// DeleteMember implements tenantpb.TenantServiceServer.
func (t *Tenant) DeleteMember(ctx context.Context,
	req *tenantpb.DeleteMemberRequest,
) (*tenantpb.DeleteMemberResponse, error) {

	zlog.Debugf("[Tenant.DeleteMember] req = %v", req)

	err := allow(ctx, tenantpb.GetTenantServiceDeleteMember())
	if err != nil {
		return nil, err
	}

	return t.svc.DeleteMember(ctx, req)
}

// allow checks that the caller is authenticated and may call method.
func allow(ctx context.Context, method *typepb.XMethod) error {
	ss, err := headers.GetAuth(ctx)
	if err != nil {
		return err
	}

	return ss.Check(method)
}
