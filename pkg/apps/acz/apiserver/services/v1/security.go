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

	grpcgateway "github.com/sentinez/core/grpc/gateway"
	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/security/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
)

var _ grpcgateway.ServiceRegistrar = (*security)(nil)

// NewSecurity creates a new security service registrar.
func NewSecurity(
	srv securitypb.SecurityServiceServer) grpcgateway.ServiceRegistrar {
	return &security{server: srv}
}

// security is the security service registrar.
type security struct {
	server securitypb.SecurityServiceServer
}

// AcceptFromEndpoint implements httpx.ServiceRegistrar.
func (s *security) AcceptFromEndpoint(ctx context.Context,
	server grpcgateway.Server, appConf *confpb.Config) error {

	return grpcgateway.RegisterServiceFromEndpoint(ctx,
		appConf,
		server.RuntimeMux(),
		securitypb.GetMetaSecurityServiceKey(),
		securitypb.RegisterSecurityServiceHandlerFromEndpoint,
	)
}

// Accept to visit the security service.
func (s *security) Accept(ctx context.Context,
	server grpcgateway.Server) error {

	return grpcgateway.RegisterServiceHandlerServer(ctx,
		server.RuntimeMux(),
		s.server,
		securitypb.RegisterSecurityServiceHandlerServer,
	)
}
