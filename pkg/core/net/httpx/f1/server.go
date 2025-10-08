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

package httpxf1

import (
	"context"
	"crypto/tls"

	"github.com/sentinez/sentinez"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/core/net/httpx"
	"github.com/sentinez/sentinez/pkg/zlog"
	"github.com/valyala/fasthttp"
)

var _ Server = (*HTTPServer)(nil)

type RequestHandler func(ctx *Context) error

type Server interface {
	httpx.Server
	Use(mdw ...func(handler RequestHandler) RequestHandler)
	Handle(fn func(ctx *Context) error)
	ListenAndServeTLS(addr, certFile, keyFile string) error
}

// NewServer creates a new fasthttp server instance.
// It implements the platform.Server interface.
func NewServer(meta *common.SntzMeta) Server {
	return &HTTPServer{
		core: &fasthttp.Server{},
		meta: meta,
	}
}

// HTTPServer implements the Server interface.
type HTTPServer struct {
	core   *fasthttp.Server
	chains []func(RequestHandler) RequestHandler
	meta   *common.SntzMeta
}

// Use implements Server.
func (s *HTTPServer) Use(mdw ...func(handler RequestHandler) RequestHandler) {
	s.chains = append(s.chains, mdw...)
}

func (s *HTTPServer) Handle(fn func(ctx *Context) error) {

	handler := func(ctx *fasthttp.RequestCtx) {
		c := NewContext(ctx)

		for i := len(s.chains) - 1; i >= 0; i-- {
			fn = s.chains[i](fn)
		}

		if err := fn(c); err != nil {
			zlog.Debugf("httpxf1: err=%v", err)
		}

		c.Release()
	}

	s.core.Handler = wrapHandler(handler)
}

// Shutdown implements platform.Server.
func (s *HTTPServer) Shutdown(_ context.Context) error {
	return s.core.Shutdown()
}

func (s *HTTPServer) initialize(addr string) {
	sentinez.INFO(
		s.meta.GetServiceName(),
		s.meta.GetServiceKey(),
	)

	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("fasthttp"),
		color.Magenta.Add(addr),
	)

	s.core.Name = sentinez.Name
	s.core.NoDefaultContentType = true
	s.core.DisableKeepalive = false
	s.core.Handler = fasthttp.CompressHandler(s.core.Handler)
}

// ListenAndServe implements platform.Server.
func (s *HTTPServer) ListenAndServe(addr string) error {

	s.initialize(addr)

	return s.core.ListenAndServe(addr)
}

func (s *HTTPServer) ListenAndServeTLS(addr, certFile, keyFile string) error {

	s.initialize(addr)

	s.core.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{"http/1.1"},
	}

	return s.core.ListenAndServeTLS(addr, certFile, keyFile)
}
