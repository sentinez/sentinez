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

// Package grpcgateway provides a http server with grpc-gateway support.
package grpcgateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"slices"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sentinez/core"
	"github.com/sentinez/core/common/console"
	corehttp "github.com/sentinez/core/http"
	httpconst "github.com/sentinez/core/http/const"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/zlog"
)

var (
	// Ensure httpServer implements HttpServer.
	_ Server = (*XServer)(nil)
)

// Server is an interface for a http server.
// default port is 9000
type Server interface {
	RuntimeMux() *runtime.ServeMux
	HTTPMux() *http.ServeMux
	Use(handlers ...func(http.Handler) http.Handler)
	ListenAndServe(addr string, opts ...corehttp.ServerOption) error
	Shutdown(ctx context.Context) error
}

// New creates a new http server.
func New(conf *settingpb.Config, opts ...runtime.ServeMuxOption) Server {
	return &XServer{
		runtimeMux: runtime.NewServeMux(opts...),
		httpMux:    http.NewServeMux(),
		meta:       conf.GetMeta(),
	}
}

func NewServer(conf *settingpb.Config) Server {
	return &XServer{
		runtimeMux: runtime.NewServeMux(),
		httpMux:    http.NewServeMux(),
		meta:       conf.GetMeta(),
	}
}

// XServer is a http server with http serve mux and grpc-gateway serve mux.
type XServer struct {
	// grpc-gateway runtime mux
	runtimeMux *runtime.ServeMux

	// http mux
	httpMux *http.ServeMux

	// middlewares for the http server
	middlewares []func(http.Handler) http.Handler

	// http server
	server *http.Server

	meta *typepb.XMeta
}

// Start implements Server.
func (h *XServer) Start(ctx context.Context) error {
	_ = ctx
	return fmt.Errorf("grpcserver: unimplemented")
}

// Use middleware for the http server. Middleware will be called
// in the order they are added, top to bottom. the middleware will
// be executed before the http handler.
func (h *XServer) Use(handlers ...func(http.Handler) http.Handler) {
	h.middlewares = append(h.middlewares, handlers...)
}

// ListenAndServe starts the runtime mux.
func (h *XServer) ListenAndServe(address string,
	opts ...corehttp.ServerOption) error {
	if address == "" {
		address = ":9000"
	}

	h.httpMux.Handle("/", h.runtimeMux)
	h.server = &http.Server{
		Addr:    address,
		Handler: chain(h.httpMux, h.middlewares...),
	}
	host, port, _ := net.SplitHostPort(address)

	var option corehttp.Option
	for _, opt := range opts {
		opt(&option)
	}

	if option.CertFile != "" && option.CertKeyFile != "" {
		if option.TLSConfig != nil {
			zlog.Warnf("grpc http: http ignore TLS config")
		}
		console.INFO(
			h.meta.GetServiceName(),
			h.meta.GetServiceKey(),
			fmt.Sprintf("running on https %s:%s", host, port),
		)
		return h.server.ListenAndServeTLS(option.CertFile, option.CertKeyFile)
	}

	console.INFO(
		h.meta.GetServiceName(),
		h.meta.GetServiceKey(),
		fmt.Sprintf("running on http %s:%s", host, port),
	)
	return h.server.ListenAndServe()
}

// RuntimeMux returns the underlying runtime mux.
func (h *XServer) RuntimeMux() *runtime.ServeMux {
	return h.runtimeMux
}

// HTTPMux returns the underlying http mux
func (h *XServer) HTTPMux() *http.ServeMux {
	return h.httpMux
}

// Shutdown implements HttpServer.
func (h *XServer) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func chain(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for _, v := range slices.Backward(m) {
		h = v(h)
	}

	return extendHeader(h)
}

func extendHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		w.Header().Set(httpconst.HeaderServer, core.Name)
	})
}
