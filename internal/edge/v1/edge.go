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
	httpxhz "github.com/sentinez/sentinez/pkg/network/httpx/hz"
	"github.com/sentinez/sentinez/pkg/runner/v1"
	"github.com/sentinez/sentinez/pkg/zlog"
)

// New creates a new Edge Server instance.
func New(server httpxhz.Server, yaml *edgeyaml.Config) *Server {
	return &Server{
		core: server,
		yaml: yaml,
	}
}

// Server implements the Edge Server interface.
// Main function and handler of the edge service.
// All traffic will be handled by this server.
type Server struct {
	core httpxhz.Server
	yaml *edgeyaml.Config
}

// Shutdown implements v1.Server.
func (s *Server) Shutdown(ctx context.Context) error {
	zlog.Debugf("application is shutting down")
	return s.core.Shutdown(ctx)
}

// Start implements v1.Server.
func (s *Server) Start(ctx context.Context) error {
	appConf := runner.GetAppConfig(ctx)
	if err := s.initialize(appConf); err != nil {
		zlog.Errorf("failed to initial: %v", err)
		return err
	}

	var (
		addr     = appConf.GetEnvConf().GetHttpAddress()
		certFile = appConf.GetFlag().GetCertificateFile()
		keyFile  = appConf.GetFlag().GetCertKeyFile()
	)

	return s.core.ListenAndServeTLS(addr, certFile, keyFile)
}
