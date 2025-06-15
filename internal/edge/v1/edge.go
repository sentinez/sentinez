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

// Package edge provides the implementation of the EdgeServiceServer interface.
package edge

import (
	"context"

	edgeconfig "github.com/sentinez/sentinez/cmd/edge/v1/apps/config"
	"github.com/sentinez/sentinez/internal/edge/v1/origin"
	"github.com/sentinez/sentinez/pkg/common/color"
	httpv2mdw "github.com/sentinez/sentinez/pkg/core/httpx/v2/middleware"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	sentinezpb "github.com/sentinez/sentinez/api/gen/go/sentinez/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/proxy"
	httpxv2 "github.com/sentinez/sentinez/pkg/core/httpx/v2"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
)

var (
	_ Edge            = (*Server)(nil)
	_ sentinez.Server = (*Server)(nil)
)

var (
	_ = sentinez.Inject(httpxv2.NewServer)
)

// Edge is the interface that wraps the basic Serve method.
type Edge interface {
	Serve(addr string) error
}

// New creates a new Edge Server instance.
func New(server httpxv2.Server,
	flag *sentinezpb.FlagEdge, conf *edgeconfig.Routes) sentinez.Server {
	return &Server{
		core:   server,
		flag:   flag,
		config: conf,
	}
}

// Server implements the Edge Server interface.
// Main function and handler of the edge service.
// All traffic will be handled by this server.
type Server struct {
	core   httpxv2.Server
	config *edgeconfig.Routes
	flag   *sentinezpb.FlagEdge
}

// Shutdown implements v1.Server.
func (s *Server) Shutdown(_ context.Context) error {
	return s.core.Shutdown()
}

// Start implements v1.Server.
func (s *Server) Start(_ context.Context) error {
	return s.Serve(s.flag.GetAddress())
}

// Serve starts the server and listens on the given address.
func (s *Server) Serve(addr string) error {
	edge.PrintASCII()

	origin.SetOrigin(s.config)

	protected := httpv2mdw.Protected(s.flag.GetRulePath())
	s.core.Use(protected)

	poolProxy, err := proxy.NewChanPool()
	if err != nil {
		return err
	}

	s.core.Handle(proxy.Handler(poolProxy, "/"))

	zlog.Infof("[%s] boost!!!", color.Blue.Add("FastHTTP"))
	zlog.Infof("[HTTP] listen on: %s", color.Magenta.Add(s.flag.GetAddress()))
	return s.core.ListenAndServe(addr)
}
