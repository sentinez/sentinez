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

package httpxv2

import (
	"github.com/sentinez/sentinez/pkg/core/httpx"
	"github.com/valyala/fasthttp"
)

var _ Server = (*server)(nil)

type Server interface {
	httpx.Server
	Use(mdw ...func(handler fasthttp.RequestHandler) fasthttp.RequestHandler)
	Handle(fn func(ctx *Context) error)
}

// NewServer creates a new fasthttp server instance.
// It implements the platform.Server interface.
func NewServer() Server {
	return &server{
		core: &fasthttp.Server{},
	}
}

// server implements the Server interface.
type server struct {
	core *fasthttp.Server
	mdw  []func(handler fasthttp.RequestHandler) fasthttp.RequestHandler
}

func (s *server) Use(
	mdw ...func(handler fasthttp.RequestHandler) fasthttp.RequestHandler) {
	s.mdw = append(s.mdw, mdw...)
}

func (s *server) Handle(fn func(ctx *Context) error) {
	handler := func(ctx *fasthttp.RequestCtx) {
		c := NewContext(ctx)
		if err := fn(c); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
	}

	// Apply middleware in reverse order (last added wraps the inner)
	for i := len(s.mdw) - 1; i >= 0; i-- {
		handler = s.mdw[i](handler)
	}

	// Apply fixed final wrapper (e.g., Server header)
	final := func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			// Set a custom Server header
			ctx.Response.Header.Set("Server", "sentinez")
			next(ctx)
		}
	}

	s.core.Handler = final(handler)
}

// Shutdown implements platform.Server.
func (s *server) Shutdown() error {
	return s.core.Shutdown()
}

// ListenAndServe implements platform.Server.
func (s *server) ListenAndServe(addr string) error {
	return s.core.ListenAndServe(addr)
}
