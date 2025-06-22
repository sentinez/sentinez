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

// Package grpcgw provides a gRPC server for the sentinez.
package grpcgw

import (
	"context"
	"fmt"

	discoverypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/discovery/v1"
	"github.com/sentinez/sentinez/pkg/common/cron"
	"github.com/sentinez/sentinez/pkg/common/protobuf"
	httpgw "github.com/sentinez/sentinez/pkg/core/gateway/http"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/sentinez/sentinez/plugins/sdk/discovery/v1"

	"google.golang.org/grpc"
)

var (
	// Ensure Server implements ServiceServer.
	_ ServiceServer = (*Server)(nil)

	// Ensure Server implements Server.
	_ sentinez.Server = (*Server)(nil)
)

// ServiceServer is a gRPC service server.
type ServiceServer interface {
	AsServer() *grpc.Server
	Serve(info *ServiceInfo) error
	Shutdown(ctx context.Context) error
}

// Server is a gRPC server that registers services.
// inherit in <Service>Server:
//
//	type Greeter struct {
//		*core.Server
//		config *types.Config
//		srv    greeter.GreeterServiceServer
//	}
type Server struct {
	server *grpc.Server
}

// Start implements Server.
func (s *Server) Start(ctx context.Context) error {
	_ = ctx
	return errors.ErrUnimplemented
}

// Shutdown implements ServiceServer.
func (s *Server) Shutdown(_ context.Context) error {
	s.server.GracefulStop()
	return nil
}

// AsServer returns the underlying gRPC server.
// return the underlying gRPC server.
func (s *Server) AsServer() *grpc.Server {
	return s.server
}

// Serve starts the http server.
// return error if the http server fails to start.
func (s *Server) Serve(info *ServiceInfo) error {
	if info == nil {
		return fmt.Errorf("grpc-service-server: info is nil")
	}

	listener, err := httpgw.ListenNetworkTCP(info.Addr)
	if err != nil {
		return err
	}

	host, port, err := httpgw.SplitHostPortListener(listener)
	if err != nil {
		return err
	}

	cron.Start(context.Background(), info.TTL, func() {
		if err := discovery.New(info.GatewayAddr).Register(context.Background(),
			&discoverypb.RegisterRequest{
				Name:    info.Name,
				Address: fmt.Sprintf("%s:%d", host, port),
				Ttl:     protobuf.ToDuration(info.TTL),
			}); err != nil {

			zlog.Errorf("grpc-service-server: error when register %v", err)
		}
	})

	return s.AsServer().Serve(listener)
}

// New returns a new service registrar.
// opts are the gRPC server options.
func New(opts ...grpc.ServerOption) *Server {
	return &Server{
		server: grpc.NewServer(opts...),
	}
}

// NewDefault returns a new service registrar with default options.
func NewDefault() *Server {
	return New()
}
