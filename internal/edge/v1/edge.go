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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	"github.com/sentinez/sentinez/pkg/client/names"
	"github.com/sentinez/sentinez/pkg/common/color"
	httpxf1 "github.com/sentinez/sentinez/pkg/core/net/httpx/f1"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/version"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	_ Edge          = (*Server)(nil)
	_ runner.Server = (*Server)(nil)
)

// Edge is the interface that wraps the basic Serve method.
type Edge interface {
	Serve(addr string) error
}

// New creates a new Edge Server instance.
func New(server httpxf1.Server,
	flag *common.FlagEdge, conf *edgeyaml.Config) runner.Server {
	return &Server{
		core:   server,
		flag:   flag,
		config: conf,
		logger: zlog.NewJSON(common.SNTZ_SNTZ_EDGE.String(), zlog.LevelWarning),
	}
}

// Server implements the Edge Server interface.
// Main function and handler of the edge service.
// All traffic will be handled by this server.
type Server struct {
	core   httpxf1.Server
	config *edgeyaml.Config
	flag   *common.FlagEdge
	logger zlog.Logger
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
//
//nolint:funlen
func (s *Server) Serve(addr string) error {
	version.ASCII("SENTINEZ // EDGE", names.EdgeV1.String())

	if err := s.bootloader(context.Background()); err != nil {
		zlog.Errorf("failed to bootloader: %v", err)
		return err
	}

	zlog.Infof("%s engine boost on: %s",
		color.Blue.Add("[fasthttp]"),
		color.Magenta.Add(s.flag.GetAddress()),
	)

	return s.core.ListenAndServe(addr)
}
