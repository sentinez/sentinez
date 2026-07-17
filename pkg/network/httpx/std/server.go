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
	"crypto/tls"
	"net/http"
	"time"

	corehttp "github.com/sentinez/core/http"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
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
}

func (s *Server) Use(
	mdw ...func(next corehttp.RequestHandler) corehttp.RequestHandler) {
	s.mdw = append(s.mdw, mdw...)
}

func (s *Server) Handle(fn corehttp.RequestHandler) {

	s.mux.Handle("/", Convert(chain(fn, s.mdw...)))
}

func (s *Server) TLS(tlsFn func(*tls.ClientHelloInfo) (*tls.Config, error)) {
	s.core.TLSConfig.GetConfigForClient = tlsFn
}

func (s *Server) ListenAndServe(
	addr string, opts ...corehttp.ServerOption) error {

	var option corehttp.Option
	for _, opt := range opts {
		opt(&option)
	}

	s.core.Addr = addr
	s.core.Handler = s.mux

	if option.CertFile != "" && option.CertKeyFile != "" {
		s.core.IdleTimeout = 120 * time.Second
		s.core.ReadTimeout = 15 * time.Second
		s.core.WriteTimeout = 15 * time.Second
		s.core.TLSConfig = option.TLSConfig

		return s.core.ListenAndServeTLS(option.CertFile, option.CertKeyFile)
	}

	return s.core.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.core.Shutdown(ctx)
}

func chain(
	h corehttp.RequestHandler,
	m ...func(corehttp.RequestHandler) corehttp.RequestHandler,
) corehttp.RequestHandler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}

	return h
}
