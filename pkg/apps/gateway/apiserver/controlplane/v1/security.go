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

package controlplane

import (
	"context"

	"github.com/sentinez/core/szgrpc/szgrpcgateway"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ szgrpcgateway.ServiceRegistrar = (*security)(nil)

// NewSecurity creates a new security service registrar.
func NewSecurity(
	srv securitypb.SecurityServiceServer) szgrpcgateway.ServiceRegistrar {
	return &security{server: srv}
}

// security is the security service registrar.
type security struct {
	server securitypb.SecurityServiceServer
}

// AcceptFromEndpoint implements httpx.ServiceRegistrar.
func (s *security) AcceptFromEndpoint(ctx context.Context,
	server szgrpcgateway.Server, appConf *settingpb.Config) error {

	return szgrpcgateway.RegisterServiceFromEndpoint(ctx,
		appConf,
		server.RuntimeMux(),
		securitypb.GetMetaSecurityServiceKey(),
		securitypb.RegisterSecurityServiceHandlerFromEndpoint,
	)
}

// Accept to visit the security service.
func (s *security) Accept(ctx context.Context,
	server szgrpcgateway.Server) error {

	// a local conn, unlike RegisterSecurityServiceHandlerServer, runs
	// interceptors, so requests are validated at the gateway.
	conn := szgrpcgateway.NewLocalConn()
	securitypb.RegisterSecurityServiceServer(conn, s.server)

	return securitypb.RegisterSecurityServiceHandlerClient(ctx,
		server.RuntimeMux(),
		securitypb.NewSecurityServiceClient(conn),
	)
}
