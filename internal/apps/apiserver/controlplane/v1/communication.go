// Copyright 2026 Sentinéz Labs.
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

	"github.com/sentinez/core/szgrpc/grpcgateway"
	communicationpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/communication/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"google.golang.org/grpc"
)

var _ grpcgateway.ServiceRegistrar = (*Communication)(nil)

// Communication registers the communication service to the gateway.
type Communication struct {
	server       communicationpb.CommunicationServiceServer
	interceptors []grpc.UnaryServerInterceptor
}

// NewCommunication creates the registrar of the communication service.
func NewCommunication(
	server communicationpb.CommunicationServiceServer,
	interceptors ...grpc.UnaryServerInterceptor,
) *Communication {
	return &Communication{server: server, interceptors: interceptors}
}

// AcceptFromEndpoint implements grpcgateway.ServiceRegistrar.
func (c *Communication) AcceptFromEndpoint(ctx context.Context,
	server grpcgateway.Server, appConf *settingpb.Config) error {
	return grpcgateway.RegisterServiceFromEndpoint(ctx,
		appConf,
		server.RuntimeMux(),
		communicationpb.GetMetaCommunicationServiceKey(),
		communicationpb.RegisterCommunicationServiceHandlerFromEndpoint,
	)
}

// Accept implements grpcgateway.ServiceRegistrar.
func (c *Communication) Accept(ctx context.Context,
	server grpcgateway.Server) error {
	// a local conn, unlike RegisterCommunicationServiceHandlerServer, runs
	// interceptors, so requests are validated at the gateway.
	conn := grpcgateway.NewLocalConn(chain(c.interceptors)...)
	communicationpb.RegisterCommunicationServiceServer(conn, c.server)

	return communicationpb.RegisterCommunicationServiceHandlerClient(ctx,
		server.RuntimeMux(),
		communicationpb.NewCommunicationServiceClient(conn),
	)
}
