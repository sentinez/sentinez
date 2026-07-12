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

package coregrpc

import (
	"context"
	"fmt"
	"net"

	"github.com/sentinez/core/common/console"
	grpcgateway "github.com/sentinez/core/grpc/gateway"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// ServiceServer is a gRPC service server.
type ServiceServer interface {
	AsServer() *grpc.Server
	Serve(conf *confpb.Config) error
	Shutdown(ctx context.Context) error
}

// Server is a gRPC server that registers services.
type Server struct {
	server *grpc.Server
	option Option
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
func (s *Server) Serve(conf *confpb.Config) error {

	addr := conf.GetEnv().GetGrpcAddress()
	listener, err := grpcgateway.ListenNetworkTCP(addr)
	if err != nil {
		return err
	}

	if s.option.consul {
		go Register(s.option.meta.GetServiceKey(), conf.GetEnv())
	}

	host, port, _ := net.SplitHostPort(addr)
	console.INFO(
		s.option.meta.GetServiceName(),
		s.option.meta.GetServiceKey(),
		fmt.Sprintf("grpc running on %s:%s", host, port),
	)

	return s.AsServer().Serve(listener)
}

func (s *Server) BufServe(bufLis *bufconn.Listener) error {
	return s.AsServer().Serve(bufLis)
}

// New returns a new service registrar.
// opts are the gRPC server options.
func New(opts ...ServerOption) *Server {

	server := &Server{}
	for _, opt := range opts {
		opt(&server.option)
	}

	server.server = grpc.NewServer(server.option.grpcOpt...)
	return server
}

// NewDefault returns a new service registrar with default options.
func NewDefault() *Server {
	return New()
}

func NewDefaultServer() *Server {
	return &Server{
		server: grpc.NewServer(),
	}
}
