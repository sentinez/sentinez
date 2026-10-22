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
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sentinez/core"
	"github.com/sentinez/core/szcmn/console"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szerror"
	"github.com/sentinez/shared/szlog"
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
	ListenAndServe(addr string, opts ...szhttp.ServerOption) error
	Shutdown(ctx context.Context) error
}

// New creates a new http server.
func New(conf *settingpb.Config, opts ...runtime.ServeMuxOption) Server {
	return &XServer{
		runtimeMux: runtime.NewServeMux(append(defaultSrvMuxOpt(), opts...)...),
		httpMux:    http.NewServeMux(),
		meta:       conf.GetMeta(),
		server:     &http.Server{},
	}
}

func NewServer(conf *settingpb.Config) Server {
	return &XServer{
		runtimeMux: runtime.NewServeMux(defaultSrvMuxOpt()...),
		httpMux:    http.NewServeMux(),
		meta:       conf.GetMeta(),
		server:     &http.Server{},
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
	opts ...szhttp.ServerOption) error {
	if address == "" {
		address = ":9000"
	}

	h.httpMux.Handle("/", h.runtimeMux)

	h.server.Addr = address
	h.server.Handler = chain(h.httpMux, h.middlewares...)

	host, port, _ := net.SplitHostPort(address)

	option := szhttp.NewOption(opts...)

	if option.CertFile != "" && option.CertKeyFile != "" {
		if option.TLSConfig != nil {
			szlog.Warnf("grpc http: http ignore TLS config")
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

func routingErrorHandler(ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	writer http.ResponseWriter,
	request *http.Request,
	status int) {

	runtime.DefaultRoutingErrorHandler(
		ctx, mux, marshaler, writer, request, status)
}

func errorHandler(ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	writer http.ResponseWriter,
	request *http.Request,
	err error) {

	switch {
	case errors.Is(err, szerror.ErrNotFound):
		err = szerror.StatusNotFoundF("%v", err)
	case errors.Is(err, szerror.ErrInvalidData):
		err = szerror.StatusInvalidDataF("%v", err)
	case errors.Is(err, szerror.ErrUnimplemented):
		err = szerror.StatusUnimplementedF("%v", err)
	}

	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, err)
}

func defaultSrvMuxOpt() []runtime.ServeMuxOption {
	return []runtime.ServeMuxOption{
		runtime.WithRoutingErrorHandler(routingErrorHandler),
		runtime.WithErrorHandler(errorHandler),
	}
}
