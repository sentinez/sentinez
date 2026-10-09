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

package stdhttpx

import (
	"context"
	"net/http"
	"slices"
	"time"

	corehttp "github.com/sentinez/core/http"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/pkg/network"
	"github.com/sentinez/sentinez/pkg/network/httpx"
)

var _ corehttp.Server = (*Server)(nil)

func NewServer(appConf *settingpb.Config) corehttp.Server {
	return corehttp.DecoreServer(appConf, &Server{
		core: &http.Server{},
		mux:  http.NewServeMux(),
	})
}

type Server struct {
	mdw  []func(corehttp.RequestHandler) corehttp.RequestHandler
	core *http.Server
	mux  *http.ServeMux
	opt  corehttp.Option
}

// AcceptReverse implements [corehttp.Server].
func (s *Server) AcceptReverse(target string) (corehttp.ReverseProxy, error) {
	return httpx.NewReverseProxy(target, network.StandardTransporter())
}

func (s *Server) Use(
	mdw ...func(next corehttp.RequestHandler) corehttp.RequestHandler) {
	s.mdw = append(s.mdw, mdw...)
}

func (s *Server) Handle(fn corehttp.RequestHandler) {

	s.mux.Handle("/", Convert(chain(fn, s.mdw...)))
}

func (s *Server) ListenAndServe(
	addr string, opts ...corehttp.ServerOption) error {
	for _, opt := range opts {
		opt(&s.opt)
	}

	s.core.Addr = addr
	s.core.Handler = s.mux

	s.onAcceptConn()

	if s.opt.CertFile != "" && s.opt.CertKeyFile != "" {
		s.core.IdleTimeout = 120 * time.Second
		s.core.ReadTimeout = 15 * time.Second
		s.core.WriteTimeout = 15 * time.Second
		s.core.TLSConfig = s.opt.TLSConfig

		return s.listenAndServeTLS(s.opt.CertFile, s.opt.CertKeyFile)
	}

	return s.listenAndServe()
}

func (s *Server) onAcceptConn() {
	if s.opt.OnStdConnect == nil {
		return
	}

	s.core.ConnContext = s.opt.OnStdConnect
}

func (s *Server) listenAndServe() error {
	if s.core.Addr == "" {
		s.core.Addr = ":http"
	}

	if s.opt.StdListener == nil {
		l, err := network.StdListen(s.core.Addr, network.WithTCP())
		if err != nil {
			return err
		}
		defer func() { _ = l.Close() }()

		s.opt.StdListener = l
	}

	return s.core.Serve(s.opt.StdListener)
}

func (s *Server) listenAndServeTLS(certFile, keyFile string) error {
	if s.core.Addr == "" {
		s.core.Addr = ":https"
	}

	if s.opt.StdListener == nil {
		l, err := network.StdListen(s.core.Addr, network.WithTCP())
		if err != nil {
			return err
		}
		defer func() { _ = l.Close() }()

		s.opt.StdListener = l
	}

	return s.core.ServeTLS(s.opt.StdListener, certFile, keyFile)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.core.Shutdown(ctx)
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
