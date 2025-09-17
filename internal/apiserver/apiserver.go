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

	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func New(server httpgw.Server) *Server {
	srv := &Server{
		server: server,
	}

	return srv
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

	appConf := runner.GetAppConfig(ctx)
	for _, service := range services {
		err := service.AcceptFromEndpoint(ctx, srv.server, appConf)
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
	if err := srv.Bootloader(ctx); err != nil {
		zlog.Errorf("apiserver: failed to bootloader: %v", err)
		return err
	}

	appConf := runner.GetAppConfig(ctx)
	// Listen HTTP server (and apiserver calls to gRPC server endpoint)
	return srv.server.Listen(appConf.GetEnvConf().GetAddress())
	// for DEBUG:
	// return errors.F("apiserver: failed to listen and serve")
}

// Shutdown implements runner.Server.
func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.server.Shutdown(ctx)
}
