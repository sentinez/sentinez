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

package szgrpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/sentinez/core/internal/console"
	"github.com/sentinez/core/szgrpc/grpcgateway"
	"github.com/sentinez/sentinez/api/interceptor/validate"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// ServiceServer is a gRPC service server.
type ServiceServer interface {
	AsServer() *grpc.Server
	Serve(conf *settingpb.Config) error
	Shutdown(ctx context.Context) error
}

// Server is a gRPC server that registers services.
type Server struct {
	server *grpc.Server
	option Option

	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Shutdown implements ServiceServer.
func (s *Server) Shutdown(_ context.Context) error {
	s.mu.Lock()
	cancel := s.cancel
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	s.server.GracefulStop()
	s.wg.Wait()

	return nil
}

// register keeps the service listening on addr registered in consul until
// Shutdown.
func (s *Server) register(conf *settingpb.Config, addr string) {
	ctx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	s.cancel = cancel
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		Register(ctx, conf.GetMeta().GetServiceKey(), addr, conf)
	}()
}

// AsServer returns the underlying gRPC server.
// return the underlying gRPC server.
func (s *Server) AsServer() *grpc.Server {
	return s.server
}

// Serve starts the http server.
// return error if the http server fails to start.
func (s *Server) Serve(conf *settingpb.Config, opts ...ServerOption) error {
	for _, opt := range opts {
		opt(&s.option)
	}

	addr := conf.Get(settingpb.Senz_SENZ_ADDRESS)
	if s.option.address != "" {
		addr = s.option.address
	}

	listener, err := grpcgateway.ListenNetworkTCP(addr)
	if err != nil {
		return err
	}

	if s.option.consul {
		s.register(conf, addr)
	}

	host, port, _ := net.SplitHostPort(addr)
	console.INFO(
		conf.GetMeta().GetServiceName(),
		conf.GetMeta().GetServiceKey(),
		fmt.Sprintf("grpc running on %s:%s", host, port),
	)

	return s.AsServer().Serve(listener)
}

func (s *Server) BufServe(bufLis *bufconn.Listener) error {
	return s.AsServer().Serve(bufLis)
}

// New returns a new service registrar.
// opts are the gRPC server options. Requests are validated against their
// buf.validate rules before any interceptor given in opts runs.
func New(opts ...grpc.ServerOption) *Server {
	opts = append([]grpc.ServerOption{
		grpc.ChainUnaryInterceptor(validate.UnaryServerInterceptor()),
	}, opts...)

	return &Server{server: grpc.NewServer(opts...)}
}

// NewDefault returns a new service registrar with default options.
func NewDefault() *Server {
	return New()
}

func NewDefaultServer() *Server {
	return New()
}
