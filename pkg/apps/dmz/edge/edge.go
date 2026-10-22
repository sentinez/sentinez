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
	"errors"
	"fmt"

	"github.com/sentinez/core/szcmn"
	"github.com/sentinez/core/szhttp"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	edgeapi "github.com/sentinez/sentinez/internal/dmz/edge/api"
	"github.com/sentinez/sentinez/internal/dmz/edge/engine"
	"github.com/sentinez/sentinez/internal/dmz/edge/transport"
	"github.com/sentinez/sentinez/internal/memory"
	edgeyaml "github.com/sentinez/sentinez/pkg/apps/dmz/edge/yaml"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/sentinez/pkg/sdklog"
	"github.com/sentinez/shared/zlog"
	otelsdklog "go.opentelemetry.io/otel/sdk/log"
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

// NewServer initializes and returns a new Edge Server instance.
// The Edge Server is responsible for handling all external HTTP traffic
//
// Parameters:
//   - server: The HTTP DMZ server implementation handling request I/O.
//   - setting: The loaded proxy configuration for routing and filtering.
//
// Returns:
//   - *Server: A new Edge Server instance ready to be started.
//   - error: error when create server
func NewServer(conf *settingpb.Config) (*Server, error) {

	localSetting, err := edgeyaml.LoadSetting(conf)
	if err != nil {
		return nil, err
	}

	szcmn.NormalizeEdgeSetting(localSetting)

	mem, err := memory.NewMemStore(conf)
	if err != nil {
		return nil, err
	}

	logExporter, err := sdklog.NewProvider(conf)
	if err != nil {
		return nil, err
	}

	return &Server{
		setting:     localSetting,
		memstore:    mem,
		conf:        conf,
		logExporter: logExporter,
		service:     edgeapi.New(),
	}, nil
}

// Server represents the core Edge Server.
// It wraps an `stdhttp.Server` for network operations
// and holds the runtime proxy configuration.
//
// The Server is the main handler of the edge service —
// all ingress traffic is processed and dispatched here.
type Server struct {
	server      szhttp.Server
	conf        *settingpb.Config
	setting     *edgepb.Setting
	memstore    *memory.MemStore
	service     *edgeapi.EdgeService
	logExporter *otelsdklog.LoggerProvider
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
	err := s.memstore.Shutdown(ctx)

	// stop serving first, so the logs of the last requests are flushed
	err = errors.Join(err,
		s.logExporter.ForceFlush(ctx),
		s.logExporter.Shutdown(ctx),
		s.server.Shutdown(ctx),
	)

	return err
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
	server, opt := newServer(s.conf)
	s.server = server

	s.initialize()

	var (
		addr = s.conf.GetDefault(
			settingpb.Senz_SENZ_ADDRESS, defaults.HTTPSAddress)

		certFile = s.conf.GetFlag().GetCertFile()
		keyFile  = s.conf.GetFlag().GetCertKeyFile()
	)

	tlsConf, err := newTLSConfig(certFile, keyFile)
	if err != nil {
		return err
	}

	stdLis, err := network.StdListen(addr, network.WithTCP())
	if err != nil {
		return err
	}
	defer func() { _ = stdLis.Close() }()

	quicLis, err := network.QuicListen(addr, network.WithTLSConfig(tlsConf))
	if err != nil {
		return err
	}
	defer func() { _ = quicLis.Close() }()

	opts := []szhttp.ServerOption{
		szhttp.WithCertificate(certFile, keyFile),
		szhttp.WithTLSConfig(tlsConf),
		szhttp.WithStdListener(stdLis),
		szhttp.WithQuicListener(quicLis),
	}

	return s.server.ListenAndServe(addr, append(opts, opt)...)
}

// newTLSConfig loads the certificate up front: the QUIC listener clones the
// config when it is created, so certificates added later are never served.
func newTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	conf := &tls.Config{
		GetConfigForClient: transport.TLSConfig,
		MinVersion:         tls.VersionTLS13,
	}

	if certFile == "" || keyFile == "" {
		return conf, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load certificate: %w", err)
	}

	conf.Certificates = []tls.Certificate{cert}

	return conf, nil
}

func newServer(
	conf *settingpb.Config) (szhttp.Server, szhttp.ServerOption) {

	// set engine for edge server
	switch conf.GetFlag().GetEngine() {
	case "hertz":
		return engine.Hertz(conf)
	case "quic":
		return engine.Quic(conf)
	case "std":
		return engine.Standard(conf)
	default:
		return engine.Standard(conf)
	}
}
