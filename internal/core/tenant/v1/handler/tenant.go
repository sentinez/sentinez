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

	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/tenant/v1"
)

var _ tenantpb.TenantServiceServer = (*Tenant)(nil)

func New() tenantpb.TenantServiceServer {
	return &Tenant{}
}

// Tenant implement tenant.TenantServiceServer
type Tenant struct{}

// SayHello implement function of tenant.TenantServiceServer
func (t *Tenant) SayHello(ctx context.Context,
	req *tenantpb.SayHelloRequest) (*tenantpb.SayHelloResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// Status implement function of tenant.TenantServiceServer
func (t *Tenant) Status(ctx context.Context,
	req *tenantpb.StatusRequest) (*tenantpb.StatusResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}
