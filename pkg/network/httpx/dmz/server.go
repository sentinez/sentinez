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

package httpxdmz

import (
	"context"
	"crypto/tls"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
	"github.com/cloudwego/hertz/pkg/network"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/hertz-contrib/http2/factory"

	"github.com/sentinez/sentinez"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/common/tlsx"
	httpxbase "github.com/sentinez/sentinez/pkg/network/httpx/base"
	"github.com/sentinez/sentinez/pkg/zlog"
)

var _ httpxbase.Server = (*XServer)(nil)

type Server interface {
	httpxbase.Server
	Use(mdw ...func(handler RequestHandler) RequestHandler)
	Handle(fn func(ctx *Context) error)
	ListenAndServeTLS(addr, certFile, keyFile string) error
}

// NewServer creates a new hertz server instance.
// It implements the platform.Server interface.
func NewServer(meta *common.XMeta) Server {
	return &XServer{
		meta: meta,
	}
}

// XServer implements the Server interface.
type XServer struct {
	chains  []func(RequestHandler) RequestHandler
	handler app.HandlerFunc
	meta    *common.XMeta
	core    *server.Hertz
}

// Use implements Server.
func (s *XServer) Use(mdw ...func(handler RequestHandler) RequestHandler) {
	s.chains = append(s.chains, mdw...)
}

func (s *XServer) Handle(fn func(ctx *Context) error) {
	handler := func(c context.Context, ctx *app.RequestContext) {
		inCtx := NewContext(c, ctx)

		for i := len(s.chains) - 1; i >= 0; i-- {
			fn = s.chains[i](fn)
		}

		if err := fn(inCtx); err != nil {
			zlog.Errorf("[httpxdmz]: internal err=%v", err)
			_ = InternalServerError(inCtx)
		}

		inCtx.Release()
	}

	s.handler = WrapHandler(handler)
}

// Shutdown implements platform.Server.
func (s *XServer) Shutdown(ctx context.Context) error {
	if s.core == nil {
		return nil
	}

	return s.core.Shutdown(ctx)
}

func (s *XServer) TLS(certFile, keyFile string) (*tls.Config, error) {
	var certificates []tls.Certificate
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}

		certificates = []tls.Certificate{cert}
	}

	return &tls.Config{
		GetConfigForClient: func(
			chi *tls.ClientHelloInfo) (*tls.Config, error) {

			zlog.Debugf("[httpxdmz][ja4] fingerprint=%s", tlsx.JA4(chi))

			return &tls.Config{
				Certificates: certificates,
				MinVersion:   tls.VersionTLS12,
				NextProtos:   []string{"h2", "http/1.1"},
			}, nil
		},
		Certificates: certificates,
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"h2", "http/1.1"},
	}, nil
}

func (s *XServer) initialize(addr string, certFile, keyFile string) error {
	if s.meta != nil {
		sentinez.INFO(s.meta.GetServiceName(), s.meta.GetServiceKey())
	}

	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("https"),
		color.Magenta.Add(addr),
	)

	hlog.SetLevel(hlog.LevelError)

	tlsConf, err := s.TLS(certFile, keyFile)
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
		server.WithTransport(func(options *config.Options) network.Transporter {
			base := standard.NewTransporter(options)
			return &Transporter{Transporter: base}
		}),
	)

	// register http2 server factory
	s.core.AddProtocol("h2", factory.NewServerFactory())

	s.core.NoRoute(s.handler)
	s.core.Name = sentinez.Name

	return nil
}

// ListenAndServe implements platform.Server.
func (s *XServer) ListenAndServe(addr string) error {

	if err := s.initialize(addr, "", ""); err != nil {
		return err
	}

	return s.core.Run()
}

func (s *XServer) ListenAndServeTLS(addr, certFile, keyFile string) error {

	if err := s.initialize(addr, certFile, keyFile); err != nil {
		return err
	}

	return s.core.Run()
}
