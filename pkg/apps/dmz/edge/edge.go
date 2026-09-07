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
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/internal/dmz/edge/transport"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/shared/zlog"
)

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
// The Edge Server is responsible for handling all external HTTP traffic
//
// Parameters:
//   - server: The HTTP DMZ server implementation handling request I/O.
//   - setting: The loaded proxy configuration for routing and filtering.
//
// Returns:
//   - *Server: A new Edge Server instance ready to be started.
func New(conf *settingpb.Config,
	setting *edgepb.Setting,
	server corehttp.Server,
) *Server {
	corecmn.NormalizeEdgeSetting(setting)
	mem := memory.NewMemStore(setting)

	return &Server{
		conf:    conf,
		core:    server,
		setting: setting,
		mem:     mem,
	}
}

// Server represents the core Edge Server.
// It wraps an `stdhttp.Server` for network operations
// and holds the runtime proxy configuration.
//
// The Server is the main handler of the edge service —
// all ingress traffic is processed and dispatched here.
type Server struct {
	core    corehttp.Server
	conf    *settingpb.Config
	setting *edgepb.Setting
	mem     *memory.MemStore
	options []corehttp.ServerOption
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
	s.mem.Shutdown(ctx)

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
// Returns:
//   - error: Any error that occurred during startup or serving.
func (s *Server) Start() error {
	if err := s.initialize(s.conf); err != nil {
		zlog.Errorf("failed to initialize: %v", err)
		return err
	}

	var (
		addr = s.conf.GetDefault(
			settingpb.Senz_SENZ_ADDRESS, defaults.HTTPSAddress)

		certFile = s.conf.GetFlag().GetCertFile()
		keyFile  = s.conf.GetFlag().GetCertKeyFile()
	)

	l, err := network.Listen(addr, network.WithTCP())
	if err != nil {
		return err
	}
	defer func() { _ = l.Close() }()

	opts := []corehttp.ServerOption{
		corehttp.WithCertificate(certFile, keyFile),
		corehttp.WithTLSConfig(&tls.Config{
			GetConfigForClient: transport.TLSConfig,
			MinVersion:         tls.VersionTLS13,
		}),
		corehttp.WithListener(l),
	}

	return s.core.ListenAndServe(addr, append(opts, s.options...)...)
}
