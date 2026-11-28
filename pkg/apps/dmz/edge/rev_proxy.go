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

// Package edge provides the implementation of the RevProxyServiceServer
// interface.
package edge

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"sync"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsdklog"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/apps/edge/engine"
	"github.com/sentinez/sentinez/internal/apps/edge/transport"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/internal/memory"
	edgeyaml "github.com/sentinez/sentinez/pkg/apps/dmz/edge/yaml"
	"github.com/sentinez/sentinez/pkg/cluster/leader"
	"github.com/sentinez/sentinez/pkg/network"
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

// NewRevProxyServer initializes and returns a new Edge Server instance.
// The Edge Server is responsible for handling all external HTTP traffic
//
// Parameters:
//   - conf: The configuration of the edge.
//   - elector: Campaigns for the leadership of the zone (Lead).
//
// Returns:
//   - *RevProxyServer: A new Edge Server instance ready to be started.
//   - error: error when create server
func NewRevProxyServer(conf *settingpb.Config, elector *leader.Elector,
) (*RevProxyServer, error) {

	localSetting, err := edgeyaml.LoadSetting(conf)
	if err != nil {
		return nil, err
	}

	mem, err := memory.NewMemStore(conf)
	if err != nil {
		return nil, err
	}

	logExporter, err := szsdklog.New(conf)
	if err != nil {
		return nil, err
	}

	return &RevProxyServer{
		setting:     localSetting,
		memstore:    mem,
		conf:        conf,
		logExporter: logExporter,
		elector:     elector,
		synced:      make(chan struct{}),
	}, nil
}

// RevProxyServer represents the core Edge RevProxyServer.
// It wraps an `stdhttp.RevProxyServer` for network operations
// and holds the runtime proxy configuration.
//
// The RevProxyServer is the main handler of the edge service —
// all ingress traffic is processed and dispatched here.
type RevProxyServer struct {
	server      szhttp.Server
	conf        *settingpb.Config
	setting     *edgepb.Setting
	memstore    *memory.MemStore
	chain       *chains.Chain
	logExporter *szsdklog.Exporter
	elector     *leader.Elector
	// synced is closed once the memstore synced with the cluster
	synced chan struct{}

	// leading counts the running Lead; leadClosed stops new ones
	leadMu     sync.Mutex
	leadClosed bool
	leading    sync.WaitGroup
}

// Shutdown gracefully stops the Edge Server.
//
// It ensures all active connections are closed and releases
// underlying resources before the application exits.
//
// This method is automatically invoked by the `runner` package
// during the service shutdown phase.
func (s *RevProxyServer) Shutdown(ctx context.Context) error {
	szlog.Debugf("application is shutting down")
	s.endLead()

	err := s.memstore.Shutdown(ctx)

	// stop serving first, so the logs of the last requests are flushed
	err = errors.Join(err,
		s.logExporter.Shutdown(ctx),
		s.server.Shutdown(ctx),
	)

	// the chain is unused once the server stopped; nil if never started
	if s.chain != nil {
		err = errors.Join(err, s.chain.Close())
	}

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
func (s *RevProxyServer) Start() error {
	server, engineOpts := newServer(s.conf)
	s.server = server

	if err := s.initialize(); err != nil {
		return err
	}

	var (
		addr = s.conf.GetDefault(
			settingpb.Sz_SZ_ADDRESS, defaults.HTTPSAddress)

		certFile = s.conf.GetFlag().GetCertFile()
		keyFile  = s.conf.GetFlag().GetCertKeyFile()
	)

	tlsConf, err := newTLSConfig(certFile, keyFile,
		s.memstore.Certificates().GetCertificate)
	if err != nil {
		return err
	}

	lis, err := listen(addr, tlsConf, s.memstore.Passthrough)
	if err != nil {
		return err
	}
	defer lis.close()

	opts := []szhttp.ServerOption{
		szhttp.WithCertificate(certFile, keyFile),
		szhttp.WithTLSConfig(tlsConf),
		szhttp.WithStdListener(lis.std),
		szhttp.WithQuicListener(lis.quic),
	}

	return s.server.ListenAndServe(addr, append(opts, engineOpts...)...)
}

// listeners the TCP and the QUIC listener of the same address.
type listeners struct {
	std  *network.StdListener
	quic *network.QuicListener
}

// listen on addr: TCP, which passes the TLS of the domains passthrough
// resolves through untouched, and QUIC, which terminates it.
func listen(addr string, tlsConf *tls.Config,
	passthrough network.PassthroughFunc) (*listeners, error) {
	std, err := network.StdListen(addr, network.WithTCP(),
		network.WithPassthrough(passthrough))
	if err != nil {
		return nil, err
	}

	quic, err := network.QuicListen(addr, network.WithTLSConfig(tlsConf))
	if err != nil {
		_ = std.Close()

		return nil, err
	}

	return &listeners{std: std, quic: quic}, nil
}

func (l *listeners) close() {
	_ = l.std.Close()
	_ = l.quic.Close()
}

// newTLSConfig serves the certificate getCert gives for the domain the
// client names, else the one of the files. The files are loaded up front:
// the QUIC listener clones the config when it is created, so only getCert,
// called on each handshake, can serve certificates added later.
func newTLSConfig(certFile, keyFile string,
	getCert func(*tls.ClientHelloInfo) (*tls.Certificate, error),
) (*tls.Config, error) {
	conf := &tls.Config{
		GetCertificate: getCert,
		MinVersion:     tls.VersionTLS13,
	}

	if certFile == "" || keyFile == "" {
		conf.GetConfigForClient = configForClient(conf, getCert)

		return conf, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load certificate: %w", err)
	}

	// every domain has a certificate then, the one of the files
	conf.Certificates = []tls.Certificate{cert}
	conf.GetConfigForClient = transport.FingerprintHello

	return conf, nil
}

// configForClient fingerprints the client (transport.FingerprintHello), and
// refuses to resume a session for a domain getCert no longer has a
// certificate for: a resumed handshake picks no certificate, so a client
// holding a session ticket would still reach a domain whose certificate
// was removed. Its full handshake then fails with unrecognized_name.
func configForClient(conf *tls.Config,
	getCert func(*tls.ClientHelloInfo) (*tls.Certificate, error),
) func(*tls.ClientHelloInfo) (*tls.Config, error) {
	noResume := conf.Clone()
	noResume.SessionTicketsDisabled = true

	return func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
		if _, err := transport.FingerprintHello(chi); err != nil {
			return nil, err
		}

		if cert, err := getCert(chi); err == nil && cert != nil {
			return nil, nil
		}

		return noResume, nil
	}
}

func newServer(
	conf *settingpb.Config) (szhttp.Server, []szhttp.ServerOption) {

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
