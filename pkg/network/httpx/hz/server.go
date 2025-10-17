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

package httpxhz

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
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/sentinez/sentinez/pkg/x/tlsx"
	"github.com/sentinez/sentinez/pkg/zlog"
)

var _ httpx.Server = (*serverx)(nil)

type Server interface {
	httpx.Server
	Use(mdw ...func(handler RequestHandler) RequestHandler)
	Handle(fn func(ctx *Context) error)
	ListenAndServeTLS(addr, certFile, keyFile string) error
}

// NewServer creates a new hertz server instance.
// It implements the platform.Server interface.
func NewServer(meta *common.SntzMeta) Server {
	return &serverx{
		meta: meta,
	}
}

// serverx implements the Server interface.
type serverx struct {
	chains []func(RequestHandler) RequestHandler
	meta   *common.SntzMeta
	hdl    app.HandlerFunc
	core   *server.Hertz
}

// Use implements Server.
func (s *serverx) Use(mdw ...func(handler RequestHandler) RequestHandler) {
	s.chains = append(s.chains, mdw...)
}

func (s *serverx) Handle(fn func(ctx *Context) error) {

	handler := func(c context.Context, ctx *app.RequestContext) {
		inCtx := NewContext(c, ctx)

		for i := len(s.chains) - 1; i >= 0; i-- {
			fn = s.chains[i](fn)
		}

		if err := fn(inCtx); err != nil {
			zlog.Errorf("[httpxhz]: internal err=%v", err)
			_ = InternalServerError(inCtx)
		}

		inCtx.Release()
	}

	s.hdl = WrapHandler(handler)
}

// Shutdown implements platform.Server.
func (s *serverx) Shutdown(ctx context.Context) error {
	if s.core == nil {
		return nil
	}

	return s.core.Shutdown(ctx)
}

func (s *serverx) TLS(certFile, keyFile string) (*tls.Config, error) {
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

			zlog.Debugf("[httpxhz][ja4] fingerprint=%s", tlsx.JA4(chi))

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

func (s *serverx) initialize(addr string, certFile, keyFile string) error {
	sentinez.INFO(s.meta.GetServiceName(), s.meta.GetServiceKey())
	zlog.Infof("server engine >>> %s", color.Magenta.Add("HERTZ"))
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

	s.core.NoRoute(s.hdl)
	s.core.Name = sentinez.Name

	return nil
}

// ListenAndServe implements platform.Server.
func (s *serverx) ListenAndServe(addr string) error {

	if err := s.initialize(addr, "", ""); err != nil {
		return err
	}

	return s.core.Run()
}

func (s *serverx) ListenAndServeTLS(addr, certFile, keyFile string) error {

	if err := s.initialize(addr, certFile, keyFile); err != nil {
		return err
	}

	return s.core.Run()
}
