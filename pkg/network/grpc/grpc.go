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

// Package netgrpc provides a gRPC server for the sentinez.
package netgrpc

import (
	"context"
	"fmt"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/internal/shared/figure"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var (
	// Ensure Server implements ServiceServer.
	_ ServiceServer = (*Server)(nil)
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
	meta   *common.XMeta
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
	listener, err := httpx.ListenNetworkTCP(addr)
	if err != nil {
		return err
	}

	figure.INFO(s.meta.GetServiceName(),
		s.meta.GetServiceKey(), fmt.Sprintf("running on http %s", addr))

	go Register(s.meta.GetServiceKey(), conf.GetEnv())
	return s.AsServer().Serve(listener)
}

func (s *Server) BufServe(bufLis *bufconn.Listener) error {
	return s.AsServer().Serve(bufLis)
}

// New returns a new service registrar.
// opts are the gRPC server options.
func New(meta *common.XMeta, opts ...grpc.ServerOption) *Server {
	return &Server{
		server: grpc.NewServer(opts...),
		meta:   meta,
	}
}

// NewDefault returns a new service registrar with default options.
func NewDefault(meta *common.XMeta) *Server {
	return New(meta)
}

func NewDefaultServer() *Server {
	return &Server{
		server: grpc.NewServer(),
	}
}
