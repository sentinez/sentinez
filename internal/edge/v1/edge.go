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

	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	httpxf1 "github.com/sentinez/sentinez/pkg/core/net/httpx/f1"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	_ Edge          = (*Server)(nil)
	_ runner.Engine = (*Server)(nil)
)

// Edge is the interface that wraps the basic Serve method.
type Edge interface {
	Serve(addr string) error
}

// New creates a new Edge Server instance.
func New(server httpxf1.Server, yaml *edgeyaml.Config) runner.Engine {
	return &Server{
		core: server,
		yaml: yaml,
	}
}

// Server implements the Edge Server interface.
// Main function and handler of the edge service.
// All traffic will be handled by this server.
type Server struct {
	core httpxf1.Server
	yaml *edgeyaml.Config
}

// Shutdown implements v1.Server.
func (s *Server) Shutdown(_ context.Context) error {
	return s.core.Shutdown()
}

// Start implements v1.Server.
func (s *Server) Start(_ context.Context) error {
	return s.Serve(s.core.GetRunnerCtx().GetConfig().GetAddress())
}

// Serve starts the server and listens on the given address.
//
//nolint:funlen
func (s *Server) Serve(addr string) error {
	if err := s.bootloader(); err != nil {
		zlog.Errorf("failed to bootloader: %v", err)
		return err
	}

	return s.core.ListenAndServe(addr)
}
