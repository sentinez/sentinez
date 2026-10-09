// Copyright 2025 Sentinéz Labs.
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

package quichttpx

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/quic-go/quic-go/http3"
	corehttp "github.com/sentinez/core/http"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
	"golang.org/x/sync/errgroup"
)

var _ corehttp.Server = (*Server)(nil)

var errNoCertificate = errors.New("quic: tls certificate required")

const (
	_idleTimeout       = 120 * time.Second
	_readHeaderTimeout = 10 * time.Second
)

func NewServer(appConf *settingpb.Config) corehttp.Server {
	return corehttp.DecoreServer(appConf, &Server{
		h3: &http3.Server{IdleTimeout: _idleTimeout},
		tcp: &http.Server{
			IdleTimeout:       _idleTimeout,
			ReadHeaderTimeout: _readHeaderTimeout,
		},
		mux: http.NewServeMux(),
	})
}

// Server serves HTTP/3 over UDP and, on the same address, HTTP/1.1 and
// HTTP/2 over TCP. Clients always reach TCP first and only switch to HTTP/3
// after seeing the Alt-Svc header, so both listeners must be served.
type Server struct {
	mdw []func(corehttp.RequestHandler) corehttp.RequestHandler
	h3  *http3.Server
	tcp *http.Server
	mux *http.ServeMux
	opt corehttp.Option
}

// AcceptReverse implements [corehttp.Server].
func (s *Server) AcceptReverse(target string) (corehttp.ReverseProxy, error) {
	// QUIC is only terminated at the edge; upstreams are plain TCP
	// backends (HTTP/1.1 or HTTP/2), so dialing them over HTTP/3 hangs.
	return httpx.NewReverseProxy(target, network.StandardTransporter())
}

func (s *Server) Use(
	mdw ...func(next corehttp.RequestHandler) corehttp.RequestHandler) {
	s.mdw = append(s.mdw, mdw...)
}

func (s *Server) Handle(fn corehttp.RequestHandler) {
	s.mux.Handle("/", stdhttpx.Convert(chain(fn, s.mdw...)))
}

func (s *Server) ListenAndServe(
	addr string, opts ...corehttp.ServerOption) error {
	for _, opt := range opts {
		opt(&s.opt)
	}

	tlsConf, err := s.tlsConfig()
	if err != nil {
		return err
	}

	s.h3.Addr, s.h3.Handler = addr, s.mux
	s.h3.TLSConfig = tlsConf
	s.h3.ConnContext = s.opt.OnQuicConnect

	s.tcp.Addr, s.tcp.Handler = addr, http.HandlerFunc(s.serveTCP)
	// http.Server mutates NextProtos, keep it off the shared config.
	s.tcp.TLSConfig = tlsConf.Clone()
	s.tcp.ConnContext = s.opt.OnStdConnect

	return s.listenAndServe(addr, tlsConf)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return errors.Join(s.h3.Shutdown(ctx), s.tcp.Shutdown(ctx))
}

func (s *Server) listenAndServe(addr string, tlsConf *tls.Config) error {
	if s.opt.QuicListener == nil {
		l, err := network.QuicListen(addr, network.WithTLSConfig(tlsConf))
		if err != nil {
			return err
		}
		defer func() { _ = l.Close() }()

		s.opt.QuicListener = l
	}

	if s.opt.StdListener == nil {
		l, err := network.StdListen(addr, network.WithTCP())
		if err != nil {
			return err
		}
		defer func() { _ = l.Close() }()

		s.opt.StdListener = l
	}

	return s.serve(s.opt.QuicListener, s.opt.StdListener)
}

// serve runs both servers and stops the other one as soon as either exits.
func (s *Server) serve(qln http3.QUICListener, sln net.Listener) error {
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error { return s.h3.ServeListener(qln) })
	g.Go(func() error { return s.tcp.ServeTLS(sln, "", "") })
	g.Go(func() error {
		<-ctx.Done()
		return errors.Join(s.h3.Close(), s.tcp.Close())
	})

	return g.Wait()
}

// serveTCP advertises HTTP/3 so clients upgrade on their next request.
func (s *Server) serveTCP(w http.ResponseWriter, r *http.Request) {
	_ = s.h3.SetQUICHeaders(w.Header())
	s.mux.ServeHTTP(w, r)
}

// tlsConfig returns the TLS config with the certificate loaded. QUIC has no
// plaintext mode, and the QUIC listener clones the config on creation, so
// the certificate must be present before any listener is created.
func (s *Server) tlsConfig() (*tls.Config, error) {
	conf := s.opt.TLSConfig
	if conf == nil {
		conf = &tls.Config{MinVersion: tls.VersionTLS13}
	}

	if s.opt.CertFile != "" && s.opt.CertKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(s.opt.CertFile, s.opt.CertKeyFile)
		if err != nil {
			return nil, err
		}

		conf.Certificates = []tls.Certificate{cert}
	}

	if len(conf.Certificates) == 0 && conf.GetCertificate == nil {
		return nil, errNoCertificate
	}

	return conf, nil
}

func chain(
	h corehttp.RequestHandler,
	m ...func(corehttp.RequestHandler) corehttp.RequestHandler,
) corehttp.RequestHandler {
	for _, v := range slices.Backward(m) {
		h = v(h)
	}

	return h
}
