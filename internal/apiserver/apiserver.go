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

// Package apiserver provides the apiserver
package apiserver

import (
	"context"

	"github.com/sentinez/sentinez/pkg/common/protobuf"
	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

// make sure apiserver implement runner.Server
// v1.runner will be start application through runner.Server interface
var _ runner.Engine = (*Server)(nil)

// New creates a new gateway app and returns a runner.Server interface.
// This constructor is based on dependency injection. When you add parameters
// (e.g., svc dcvrhandler.Discovery), you must use the runner.Inject
// to inject the constructor of the object into the sentinez framework.
//
// Example:
//
//	var _ = runner.Inject(dcvrhandler.New)
func New(server httpgw.Server) (runner.Engine, error) {
	srv := &Server{
		server: server,
	}

	return srv, nil
}

// Server represents the sentinez app The apiserver application is the main
// entry point for the sentinez. It will automatically connect to other
// services via gRPC. Run the application along with other services
// in the cmd/ directory. The application provides APIs for users through
// a single HTTP gateway following the REST API standard. The application
// uses gRPC to connect to other services. Additionally, the system provides
// a Swagger UI interface for users to easily interact with the system
// through a web interface.
type Server struct {
	// server is the core server, manage http.ServeMux,
	// runtime.ServeMux and HTTP server
	server httpgw.Server
}

// visitToEndpoint all service to external grpc server
func (srv *Server) visitToEndpoint(ctx context.Context,
	services ...httpgw.ServiceRegistrar) error {

	for _, service := range services {
		err := service.AcceptFromEndpoint(
			ctx, srv.server, srv.server.GetConfig())
		if err != nil {
			return err
		}
	}
	return nil
	// return errors.F("apiserver: failed to visit service")
}

// visit all service to internal grpc handler
func (srv *Server) visit(ctx context.Context,
	services ...httpgw.ServiceRegistrar) error {

	for _, service := range services {
		if err := service.Accept(ctx, srv.server); err != nil {
			return err
		}
	}

	return nil
	// return errors.F("apiserver: failed to visit service")
}

// Start the apiserver/gateway app
func (srv *Server) Start(ctx context.Context) error {
	if err := protobuf.Validate(srv.server.GetConfig()); err != nil {
		return err
	}

	if err := srv.bootloader(ctx); err != nil {
		zlog.Errorf("apiserver: failed to bootloader: %v", err)
		return err
	}

	// Listen HTTP server (and apiserver calls to gRPC server endpoint)
	return srv.server.Listen(srv.server.GetConfig().GetAddress())
	// for DEBUG:
	// return errors.F("apiserver: failed to listen and serve")
}

// Shutdown implements runner.Server.
func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.server.Shutdown(ctx)
}
