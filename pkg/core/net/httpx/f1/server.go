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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/core/net/httpx"
	"github.com/sentinez/sentinez/pkg/std/version"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/valyala/fasthttp"
)

var _ Server = (*HTTPServer)(nil)

type Server interface {
	httpx.Server
	Use(mdw ...func(handler fasthttp.RequestHandler) fasthttp.RequestHandler)
	Handle(fn func(ctx *Context) error)
}

// NewHTTPServer creates a new fasthttp server instance.
// It implements the platform.Server interface.
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		core: &fasthttp.Server{},
	}
}

// HTTPServer implements the Server interface.
type HTTPServer struct {
	core     *fasthttp.Server
	mdw      []func(handler fasthttp.RequestHandler) fasthttp.RequestHandler
	Metadata *common.SentinezMetadata
}

func (s *HTTPServer) Use(
	mdw ...func(handler fasthttp.RequestHandler) fasthttp.RequestHandler) {
	s.mdw = append(s.mdw, mdw...)
}

func (s *HTTPServer) Handle(fn func(ctx *Context) error) {

	handler := func(ctx *fasthttp.RequestCtx) {
		c := NewContext(ctx)

		if err := fn(c); err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		}
	}

	// Apply middleware in reverse order (last added wraps the inner)
	for i := len(s.mdw) - 1; i >= 0; i-- {
		handler = s.mdw[i](handler)
	}

	s.core.Handler = wrapHandler(handler)
}

// Shutdown implements platform.Server.
func (s *HTTPServer) Shutdown() error {
	return s.core.Shutdown()
}

// ListenAndServe implements platform.Server.
func (s *HTTPServer) ListenAndServe(addr string) error {
	version.INFO(s.Metadata.GetServiceName(), s.Metadata.GetServiceKey())
	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("fasthttp"),
		color.Magenta.Add(addr),
	)

	s.core.Name = version.Name
	s.core.Handler = fasthttp.CompressHandler(s.core.Handler)
	return s.core.ListenAndServe(addr)
}
