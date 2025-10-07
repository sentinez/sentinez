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
package grpc

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	httpgw "github.com/sentinez/sentinez/pkg/core/net/httpx/gw"
	"github.com/sentinez/sentinez/pkg/stdcmn/zerrors"
	"github.com/sentinez/sentinez/pkg/stdcmn/zlog"
	"github.com/sentinez/sentinez/pkg/stdcmn/zversion"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const BufSize = 1024 * 1024

var (
	// Ensure Server implements ServiceServer.
	_ ServiceServer = (*Server)(nil)
)

// ServiceServer is a gRPC service server.
type ServiceServer interface {
	AsServer() *grpc.Server
	Serve(conf *common.AppConfig) error
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
	meta   *common.SntzMeta
}

// Start implements Server.
func (s *Server) Start(ctx context.Context) error {
	_ = ctx
	return zerrors.ErrUnimplemented
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
func (s *Server) Serve(conf *common.AppConfig) error {

	listener, err := httpgw.ListenNetworkTCP(conf.GetEnvConf().GetAddress())
	if err != nil {
		return err
	}

	zversion.INFO(
		s.meta.GetServiceName(),
		s.meta.GetServiceKey(),
	)

	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("gRPC"),
		color.Magenta.Add(conf.GetEnvConf().GetAddress()),
	)

	go Register(s.meta.GetServiceKey(), conf.GetEnvConf())
	return s.AsServer().Serve(listener)
}

func (s *Server) BufServe(bufLis *bufconn.Listener) error {
	return s.AsServer().Serve(bufLis)
}

// New returns a new service registrar.
// opts are the gRPC server options.
func New(meta *common.SntzMeta, opts ...grpc.ServerOption) *Server {
	return &Server{
		server: grpc.NewServer(opts...),
		meta:   meta,
	}
}

// NewDefault returns a new service registrar with default options.
func NewDefault(meta *common.SntzMeta) *Server {
	return New(meta)
}

func NewDefaultServer() *Server {
	return &Server{
		server: grpc.NewServer(),
	}
}
