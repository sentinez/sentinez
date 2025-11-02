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

	"github.com/sentinez/sentinez"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	httpgw "github.com/sentinez/sentinez/pkg/network/httpx/gw"
	"github.com/sentinez/sentinez/pkg/x/errorx"
	"github.com/sentinez/sentinez/pkg/zlog"
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
	Serve(conf *configspb.AppConfig) error
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
	meta   *common.XMeta
}

// Start implements Server.
func (s *Server) Start(ctx context.Context) error {
	_ = ctx
	return errorx.ErrUnimplemented
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
func (s *Server) Serve(conf *configspb.AppConfig) error {

	listener, err := httpgw.ListenNetworkTCP(conf.GetEnvConf().GetGrpcAddress())
	if err != nil {
		return err
	}

	sentinez.INFO(
		s.meta.GetServiceName(),
		s.meta.GetServiceKey(),
	)

	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("gRPC"),
		color.Magenta.Add(conf.GetEnvConf().GetGrpcAddress()),
	)

	go Register(s.meta.GetServiceKey(), conf.GetEnvConf())
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
