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

package controlplane

import (
	"context"

	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/core/szgrpc/grpcgateway"
	analyticpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var _ grpcgateway.ServiceRegistrar = (*analytic)(nil)

// NewAnalytic creates a new analytic service registrar.
func NewAnalytic(
	srv analyticpb.AnalyticServiceServer) grpcgateway.ServiceRegistrar {
	return &analytic{server: srv}
}

// analytic is the analytic service registrar.
type analytic struct {
	server analyticpb.AnalyticServiceServer
}

// AcceptFromEndpoint implements httpx.ServiceRegistrar.
func (a *analytic) AcceptFromEndpoint(ctx context.Context,
	server grpcgateway.Server, appConf *settingpb.Config) error {

	return grpcgateway.RegisterServiceFromEndpoint(ctx,
		appConf,
		server.RuntimeMux(),
		analyticpb.GetMetaAnalyticServiceKey(),
		analyticpb.RegisterAnalyticServiceHandlerFromEndpoint,
	)
}

// Accept to visit the analytic service.
func (a *analytic) Accept(ctx context.Context,
	server grpcgateway.Server) error {

	// a local conn, unlike RegisterAnalyticServiceHandlerServer, runs
	// interceptors, so requests are validated at the gateway.
	conn := grpcgateway.NewLocalConn(authz.Unary())
	analyticpb.RegisterAnalyticServiceServer(conn, a.server)

	return analyticpb.RegisterAnalyticServiceHandlerClient(ctx,
		server.RuntimeMux(),
		analyticpb.NewAnalyticServiceClient(conn),
	)
}
