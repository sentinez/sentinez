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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	sentinezpb "github.com/sentinez/sentinez/api/gen/go/sentinez/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/proxy"
	"github.com/sentinez/sentinez/pkg/common/color"
	httpxv1 "github.com/sentinez/sentinez/pkg/core/httpx/v1"
	"github.com/sentinez/sentinez/pkg/core/secure"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	_ Edge            = (*Server)(nil)
	_ sentinez.Server = (*Server)(nil)
)

var (
	_ = sentinez.Inject(httpxv1.NewServer)
)

// Edge is the interface that wraps the basic Serve method.
type Edge interface {
	Serve(addr string) error
}

// New creates a new Edge Server instance.
func New(server httpxv1.Server, flag *sentinezpb.FlagEdge) sentinez.Server {
	return &Server{
		core: server,
		flag: flag,
	}
}

// Server implements the Edge Server interface.
// Main function and handler of the edge service.
// All traffic will be handled by this server.
type Server struct {
	core httpxv1.Server
	flag *sentinezpb.FlagEdge
}

// Shutdown implements v1.Server.
func (s *Server) Shutdown(_ context.Context) error {
	return s.core.Shutdown()
}

// Start implements v1.Server.
func (s *Server) Start(_ context.Context) error {
	return s.Serve(s.flag.GetAddress())
}

func (s *Server) handler(ctx httpxv1.Context) error {
	return chainServe(ctx,
		proxy.Proxy,
	)
}

// Serve starts the server and listens on the given address.
func (s *Server) Serve(addr string) error {
	edge.PrintASCII()
	zlog.Infof("[HTTP] LISTEN: %s", color.Magenta.Add(s.flag.GetAddress()))

	s.core.Use(secure.ProtectServerH1)

	s.core.Handle(s.handler)

	return s.core.ListenAndServe(addr)
}
