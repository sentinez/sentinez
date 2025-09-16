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

package services

import (
	"context"

	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/tenant/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
)

var _ httpgw.ServiceRegistrar = (*tenant)(nil)

// NewTenant creates a new tenant service registrar.
func NewTenant(srv tenantpb.TenantServiceServer) httpgw.ServiceRegistrar {
	return &tenant{server: srv}
}

// tenant is the tenant service registrar.
type tenant struct {
	server tenantpb.TenantServiceServer
}

// AcceptFromEndpoint implements httpgw.ServiceRegistrar.
func (t *tenant) AcceptFromEndpoint(ctx context.Context,
	server httpgw.Server, rctx *common.RunnerCtx) error {

	return httpgw.RegisterServiceFromEndpoint(ctx,
		rctx,
		server.RuntimeMux(),
		tenantpb.GetMetaTenantServiceKey(),
		tenantpb.RegisterTenantServiceHandlerFromEndpoint,
	)
}

// Accept to visit the tenant service.
func (t *tenant) Accept(ctx context.Context, server httpgw.Server) error {

	return httpgw.RegisterServiceHandlerServer(ctx,
		server.RuntimeMux(),
		t.server,
		tenantpb.RegisterTenantServiceHandlerServer,
	)
}
