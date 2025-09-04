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

// Package httpgw provides a http server with grpc-gateway support.
package httpgw

import (
	"context"
	"net/http"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/version"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

var (
	// Ensure httpServer implements Server.
	_ runner.Engine = (*HTTPServer)(nil)

	// Ensure httpServer implements HttpServer.
	_ Server = (*HTTPServer)(nil)
)

// Server is an interface for a http server.
// default port is 9000
type Server interface {
	Listen(address string) error
	Shutdown(ctx context.Context) error
	RuntimeMux() *runtime.ServeMux
	HTTPMux() *http.ServeMux
	Use(handlers ...func(http.Handler) http.Handler)
	GetRunnerCtx() *common.RunnerCtx
}

// New creates a new http server.
func New(runnerCtx *common.RunnerCtx, opts ...runtime.ServeMuxOption) Server {
	return &HTTPServer{
		runtimeMux: runtime.NewServeMux(opts...),
		httpMux:    http.NewServeMux(),
		rctx:       runnerCtx,
	}
}

func NewServer(runnerCtx *common.RunnerCtx) Server {
	return &HTTPServer{
		runtimeMux: runtime.NewServeMux(),
		httpMux:    http.NewServeMux(),
		rctx:       runnerCtx,
	}
}

// HTTPServer is a http server with http serve mux and grpc-gateway serve mux.
type HTTPServer struct {
	// grpc-gateway runtime mux
	runtimeMux *runtime.ServeMux

	// http mux
	httpMux *http.ServeMux

	// middlewares for the http server
	middlewares []func(http.Handler) http.Handler

	// http server
	server *http.Server

	rctx *common.RunnerCtx
}

// GetRunnerCtx implements Server.
func (h *HTTPServer) GetRunnerCtx() *common.RunnerCtx {
	return h.rctx
}

// Start implements Server.
func (h *HTTPServer) Start(ctx context.Context) error {
	_ = ctx
	return errors.ErrUnimplemented
}

// Use middleware for the http server. Middleware will be called
// in the order they are added, top to bottom. the middleware will
// be executed before the http handler.
func (h *HTTPServer) Use(handlers ...func(http.Handler) http.Handler) {
	h.middlewares = append(h.middlewares, handlers...)
}

// Listen starts the runtime mux.
func (h *HTTPServer) Listen(address string) error {
	if address == "" {
		address = ":9000"
	}

	// handler runtime.Mux with http.ServeMux
	// serve grpc-gateway mux on the root path
	h.httpMux.Handle("/", h.runtimeMux)

	// create http server with address and http.Handler
	// httpMux was wrapped with the middlewares
	h.server = &http.Server{
		Addr:    address,
		Handler: chain(h.httpMux, h.middlewares...),
	}

	version.INFO(
		h.rctx.GetMeta().GetServiceName(),
		h.rctx.GetMeta().GetServiceKey(),
	)

	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("http"),
		color.Magenta.Add(address),
	)
	return h.server.ListenAndServe()
}

// RuntimeMux returns the underlying runtime mux.
func (h *HTTPServer) RuntimeMux() *runtime.ServeMux {
	return h.runtimeMux
}

// HTTPMux returns the underlying http mux
func (h *HTTPServer) HTTPMux() *http.ServeMux {
	return h.httpMux
}

// Shutdown implements HttpServer.
func (h *HTTPServer) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func chain(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}

	return extendHeader(h)
}

func extendHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		w.Header().Set("Server", version.Name)
	})
}
