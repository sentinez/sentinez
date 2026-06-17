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
	"crypto/tls"

	corecmn "github.com/sentinez/core/common"
	corehttp "github.com/sentinez/core/http"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/internal/stream"
	"github.com/sentinez/shared/zlog"
)

type ReverseProxyConstructor func(string) (corehttp.ReverseProxy, error)

//
// Package edge implements the core Edge Server component.
//
// The Edge Server acts as the main HTTP entrypoint of the system,
// handling incoming traffic and routing it through the configured
// proxy, WAF, and routing layers.
//
// This component integrates tightly with the `runner` package for
// controlled startup and graceful shutdown.
//

// New initializes and returns a new Edge Server instance.
//
// The Edge Server is responsible for handling all external HTTP traffic,
// using the provided `stdhttp.Server` as its underlying HTTP layer,
// and a given proxy `setting` configuration to determine routing,
// security, and behavior policies.
//
// Parameters:
//   - server: The HTTP DMZ server implementation handling request I/O.
//   - setting: The loaded proxy configuration for routing and filtering.
//
// Returns:
//   - *Server: A new Edge Server instance ready to be started.
func New(conf *confpb.Config,
	setting *edgepb.Setting,
	server corehttp.Server,
) *Server {
	corecmn.NormalizeEdgeSetting(setting)

	return &Server{
		conf:    conf,
		core:    server,
		setting: setting,
	}
}

// Server represents the core Edge Server.
// It wraps an `stdhttp.Server` for network operations
// and holds the runtime proxy configuration.
//
// The Server is the main handler of the edge service —
// all ingress traffic is processed and dispatched here.
type Server struct {
	conf    *confpb.Config
	core    corehttp.Server
	setting *edgepb.Setting
}

// Shutdown gracefully stops the Edge Server.
//
// It ensures all active connections are closed and releases
// underlying resources before the application exits.
//
// This method is automatically invoked by the `runner` package
// during the service shutdown phase.
func (s *Server) Shutdown(ctx context.Context) error {
	zlog.Debugf("application is shutting down")

	if err := stream.Close(); err != nil {
		zlog.Errorf("failed to close stream: %v", err)
	}

	return s.core.Shutdown(ctx)
}

// Start begins serving incoming HTTP (or HTTPS) traffic.
//
// The method initializes runtime configuration from the application context,
// prepares TLS if certificates are provided, and delegates
// the serving process to the underlying `stdhttp.Server`.
//
// This method should always be invoked through the `runner` lifecycle manager.
//
// Parameters:
//   - ctx: The lifecycle context provided by the runner.
//   - conf: application configuration
//
// Returns:
//   - error: Any error that occurred during startup or serving.
func (s *Server) Start() error {
	if err := s.initialize(s.conf); err != nil {
		zlog.Errorf("failed to initialize: %v", err)
		return err
	}

	var (
		addr     = s.conf.GetEnv().GetHttpAddress()
		certFile = s.conf.GetFlag().GetCertFile()
		keyFile  = s.conf.GetFlag().GetCertKeyFile()
	)

	return s.core.ListenAndServe(addr,
		corehttp.WithCertificate(certFile, keyFile),
		corehttp.WithTLSConfig(&tls.Config{
			GetConfigForClient: stream.TLSConfig,
		}),
	)
}
