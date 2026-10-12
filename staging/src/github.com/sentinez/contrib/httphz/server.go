// Copyright 2026 Duc-Hung Ho.
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

package httphz

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
	"github.com/cloudwego/hertz/pkg/network"
	"github.com/hertz-contrib/http2/factory"
	httphznet "github.com/sentinez/contrib/httphz/net"
	netstd "github.com/sentinez/contrib/httphz/net/std"
	proxyhz "github.com/sentinez/contrib/httphz/proxy"
	"github.com/sentinez/core"
	corehttp "github.com/sentinez/core/http"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/zlog"
)

var (
	_ corehttp.Server = (*XServer)(nil)
)

// NewServer creates a new hertz server instance.
// It implements the platform.Server interface.
func NewServer(conf *settingpb.Config) corehttp.Server {
	return corehttp.DecoreServer(conf, &XServer{})
}

// XServer implements the Server interface.
type XServer struct {
	chains  []func(corehttp.RequestHandler) corehttp.RequestHandler
	handler app.HandlerFunc
	core    *server.Hertz
}

// AcceptReverse implements [corehttp.Server].
func (s *XServer) AcceptReverse(target string) (corehttp.ReverseProxy, error) {
	return proxyhz.NewReverseProxy(target)
}

// Use implements Server.
func (s *XServer) Use(
	mdw ...func(handler corehttp.RequestHandler) corehttp.RequestHandler) {
	s.chains = append(s.chains, mdw...)
}

func (s *XServer) Handle(fn corehttp.RequestHandler) {
	handler := func(c context.Context, ctx *app.RequestContext) {
		inCtx := NewContext(c, ctx)

		for i := len(s.chains) - 1; i >= 0; i-- {
			fn = s.chains[i](fn)
		}

		if err := fn(inCtx); err != nil {
			zlog.Errorf("[httpxdmz]: internal err=%v", err)
			_ = inCtx.String(
				http.StatusInternalServerError, []byte("Internal server error"))
		}

		_ = inCtx.Close()
	}

	s.handler = WrapHandler(handler)
	// s.handler = handler
}

// Shutdown implements platform.Server.
func (s *XServer) Shutdown(ctx context.Context) error {
	if s.core == nil {
		return nil
	}

	return s.core.Shutdown(ctx)
}

func (s *XServer) TLS(
	certFile, keyFile string, tlsConfig *tls.Config) (*tls.Config, error) {

	var certificates []tls.Certificate
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}

		certificates = []tls.Certificate{cert}
	}

	if len(tlsConfig.Certificates) == 0 {
		tlsConfig.Certificates = certificates
	}

	if tlsConfig.MinVersion == 0 {
		tlsConfig.MinVersion = tls.VersionTLS12
	}

	if len(tlsConfig.NextProtos) == 0 {
		tlsConfig.NextProtos = []string{"h2", "http/1.1"}
	}

	return tlsConfig, nil
}

func (s *XServer) initialize(addr string, opt *corehttp.Option) error {
	hlog.SetLevel(hlog.LevelError)

	tlsConf, err := s.TLS(opt.CertFile, opt.CertKeyFile, opt.TLSConfig)
	if err != nil {
		return err
	}

	s.core = server.Default(
		server.WithHostPorts(addr),
		server.WithTLS(tlsConf),
		server.WithStreamBody(true),
		server.WithTraceLevel(stats.LevelDisabled),
		server.WithALPN(true),
		server.WithH2C(true),
		server.WithListener(opt.StdListener),
		server.WithOnConnect(
			func(ctx context.Context, conn network.Conn) context.Context {
				return opt.OnStdConnect(ctx, conn)
			},
		),
		server.WithTransport(func(options *config.Options) network.Transporter {
			base := netstd.NewTransporter(options)
			return &httphznet.Transporter{Transporter: base}
		}),
	)

	// register http2 server factory
	s.core.AddProtocol("h2", factory.NewServerFactory())

	s.core.NoRoute(s.handler)
	s.core.Name = core.Name

	return nil
}

// ListenAndServe implements platform.Server.
func (s *XServer) ListenAndServe(
	addr string, opts ...corehttp.ServerOption) error {

	var option corehttp.Option
	for _, opt := range opts {
		opt(&option)
	}

	if err := s.initialize(addr, &option); err != nil {
		return err
	}

	return s.core.Run()
}
