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

	sentinezpb "github.com/sentinez/sentinez/api/gen/go/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/std/names"
	"github.com/sentinez/sentinez/pkg/std/version"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

// make sure apiserver implement sentinez.Server
// v1.runner will be start application through sentinez.Server interface
var _ sentinez.Server = (*Server)(nil)

// New creates a new gateway app and returns a sentinez.Server interface.
// This constructor is based on dependency injection. When you add parameters
// (e.g., svc dcvrhandler.Discovery), you must use the sentinez.Inject
// to inject the constructor of the object into the sentinez framework.
//
// Example:
//
//	var _ = sentinez.Inject(dcvrhandler.New)
func New(conf *sentinezpb.Config,
	flag *sentinezpb.FlagAPIServer) (sentinez.Server, error) {

	srv := &Server{
		server: httpgw.New(),
		config: conf,
		flag:   flag,
	}

	if err := srv.bootloader(context.Background()); err != nil {
		zlog.Errorf("apiserver: failed to bootloader: %v", err)
		return nil, err
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
	// config is the configuration of the apiserver app, load environment
	// variables from .env file
	config *sentinezpb.Config

	// server is the core server, manage http.ServeMux,
	// runtime.ServeMux and HTTP server
	server httpgw.Server

	// flag option for the apiserver
	flag *sentinezpb.FlagAPIServer
}

// visitToEndpoint all service to external grpc server
func (srv *Server) visitToEndpoint(ctx context.Context,
	services ...httpgw.ServiceRegistrar) error {

	for _, service := range services {
		if err := service.AcceptFromEndpoint(ctx, srv.server); err != nil {
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
func (srv *Server) Start(_ context.Context) error {
	// service ascii art banner
	version.ASCII("SENTINEZ // API SERVER", names.APIServer.String())
	if err := protobuf.Validate(srv.config); err != nil {
		return err
	}

	// Listen HTTP server (and apiserver calls to gRPC server endpoint)
	// log info in console and return register error if they exist
	zlog.Infof("[HTTP] starting server %s", srv.flag.GetAddress())
	return srv.server.Listen(srv.flag.GetAddress())
	// for DEBUG:
	// return errors.F("apiserver: failed to listen and serve")
}

// Shutdown implements sentinez.Server.
func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.server.Shutdown(ctx)
}
