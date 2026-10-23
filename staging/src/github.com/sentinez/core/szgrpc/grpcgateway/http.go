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
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/quic-go/quic-go/http3"
	"github.com/sentinez/core"
	"github.com/sentinez/core/szcmn/console"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szerror"
	"golang.org/x/sync/errgroup"
)

var _ Server = (*XServer)(nil)

var errNoCertificate = errors.New("grpcgateway: tls certificate required")

const (
	// _defaultAddress is listened on when ListenAndServe gets no address.
	_defaultAddress = ":9000"
	_idleTimeout    = 120 * time.Second
)

// Server is a http server serving the grpc-gateway runtime mux next to
// custom http handlers.
type Server interface {
	RuntimeMux() *runtime.ServeMux
	HTTPMux() *http.ServeMux
	Use(handlers ...func(http.Handler) http.Handler)
	ListenAndServe(addr string, opts ...szhttp.ServerOption) error
	Shutdown(ctx context.Context) error
}

// New creates a new http server.
func New(conf *settingpb.Config, opts ...runtime.ServeMuxOption) Server {
	return newXServer(conf, append(defaultSrvMuxOpt(), opts...)...)
}

func NewServer(conf *settingpb.Config) Server {
	return newXServer(conf, defaultSrvMuxOpt()...)
}

func newXServer(
	conf *settingpb.Config, opts ...runtime.ServeMuxOption) *XServer {

	return &XServer{
		runtimeMux: runtime.NewServeMux(opts...),
		httpMux:    http.NewServeMux(),
		meta:       conf.GetMeta(),
		server:     &http.Server{IdleTimeout: _idleTimeout},
		h3:         &http3.Server{IdleTimeout: _idleTimeout},
	}
}

// XServer is a http server with http serve mux and grpc-gateway serve mux.
// With a TLS certificate it serves HTTPS over TCP and HTTP/3 over UDP on
// the same address; without one, plain HTTP over TCP.
type XServer struct {
	// grpc-gateway runtime mux
	runtimeMux *runtime.ServeMux

	// http mux
	httpMux *http.ServeMux

	// middlewares for the http server
	middlewares []func(http.Handler) http.Handler

	// server serves TCP, h3 serves HTTP/3 when TLS is configured
	server *http.Server
	h3     *http3.Server

	meta *typepb.XMeta
}

// Use middleware for the http server. Middleware will be called
// in the order they are added, top to bottom. the middleware will
// be executed before the http handler.
func (h *XServer) Use(handlers ...func(http.Handler) http.Handler) {
	h.middlewares = append(h.middlewares, handlers...)
}

// ListenAndServe serves on address (_defaultAddress when empty). The
// listeners and TLS config come from opts when set (WithStdListener,
// WithQuicListener, WithTLSConfig); a certificate (WithCertificate or in
// the TLS config) enables HTTPS and HTTP/3.
func (h *XServer) ListenAndServe(address string,
	opts ...szhttp.ServerOption) error {
	if address == "" {
		address = _defaultAddress
	}

	option := szhttp.NewOption(opts...)

	h.httpMux.Handle("/", h.runtimeMux)
	handler := chain(h.httpMux, h.middlewares...)

	h.server.Addr = address
	h.server.Handler = handler
	h.server.ConnContext = option.OnStdConnect

	tlsConf, err := tlsConfig(option)
	if errors.Is(err, errNoCertificate) {
		h.banner("http", address)
		return h.serveHTTP(option.StdListener)
	}
	if err != nil {
		return err
	}

	h.h3.Addr, h.h3.Handler = address, handler
	h.h3.TLSConfig = tlsConf
	h.h3.ConnContext = option.OnQuicConnect

	h.server.Handler = h.advertiseHTTP3(handler)
	// http.Server mutates NextProtos, keep it off the shared config.
	h.server.TLSConfig = tlsConf.Clone()

	h.banner("https and http/3", address)
	return h.serveTLS(option.StdListener, option.QuicListener)
}

// RuntimeMux returns the underlying runtime mux.
func (h *XServer) RuntimeMux() *runtime.ServeMux {
	return h.runtimeMux
}

// HTTPMux returns the underlying http mux
func (h *XServer) HTTPMux() *http.ServeMux {
	return h.httpMux
}

// Shutdown gracefully stops the TCP and HTTP/3 servers.
func (h *XServer) Shutdown(ctx context.Context) error {
	return errors.Join(h.server.Shutdown(ctx), h.h3.Shutdown(ctx))
}

func (h *XServer) banner(scheme, address string) {
	host, port, _ := net.SplitHostPort(address)
	console.INFO(
		h.meta.GetServiceName(),
		h.meta.GetServiceKey(),
		fmt.Sprintf("running on %s %s:%s", scheme, host, port),
	)
}

// serveHTTP serves plain HTTP on ln, or on a new TCP listener when nil.
func (h *XServer) serveHTTP(ln net.Listener) error {
	if ln == nil {
		return h.server.ListenAndServe()
	}

	return h.server.Serve(ln)
}

// serveTLS serves HTTPS on sln and HTTP/3 on qln, each created on the
// server address when nil, and stops both as soon as either exits.
func (h *XServer) serveTLS(sln net.Listener, qln http3.QUICListener) error {
	if sln == nil {
		ln, err := net.Listen("tcp", h.server.Addr)
		if err != nil {
			return err
		}
		defer func() { _ = ln.Close() }()

		sln = ln
	}

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error { return h.server.ServeTLS(sln, "", "") })
	g.Go(func() error {
		if qln == nil {
			return h.h3.ListenAndServe()
		}

		return h.h3.ServeListener(qln)
	})
	g.Go(func() error {
		<-ctx.Done()
		return errors.Join(h.server.Close(), h.h3.Close())
	})

	return g.Wait()
}

// advertiseHTTP3 sets Alt-Svc on TCP responses so clients switch to
// HTTP/3 on their next request.
func (h *XServer) advertiseHTTP3(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = h.h3.SetQUICHeaders(w.Header())
		next.ServeHTTP(w, r)
	})
}

// tlsConfig returns a copy of the option TLS config (TLS 1.3 when unset)
// with the option certificate loaded; errNoCertificate when there is no
// certificate to serve TLS with.
func tlsConfig(option *szhttp.Option) (*tls.Config, error) {
	conf := &tls.Config{MinVersion: tls.VersionTLS13}
	if option.TLSConfig != nil {
		conf = option.TLSConfig.Clone()
	}

	if option.CertFile != "" && option.CertKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(
			option.CertFile, option.CertKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load certificate: %w", err)
		}

		conf.Certificates = []tls.Certificate{cert}
	}

	if len(conf.Certificates) == 0 && conf.GetCertificate == nil {
		return nil, errNoCertificate
	}

	return conf, nil
}

func chain(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for _, v := range slices.Backward(m) {
		h = v(h)
	}

	return extendHeader(h)
}

// extendHeader sets the Server header before next writes the response;
// headers set after it are not sent.
func extendHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpconst.HeaderServer, core.Name)

		next.ServeHTTP(w, r)
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
