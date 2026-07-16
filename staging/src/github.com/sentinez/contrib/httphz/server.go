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
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/hertz-contrib/http2/factory"

	"github.com/sentinez/core"
	corehttp "github.com/sentinez/core/http"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
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

		inCtx.Release()
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

func (s *XServer) initialize(addr string,
	certFile, keyFile string, tlsConfig *tls.Config) error {
	hlog.SetLevel(hlog.LevelError)

	tlsConf, err := s.TLS(certFile, keyFile, tlsConfig)
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

	if err := s.initialize(addr,
		option.CertFile,
		option.CertKeyFile,
		option.TLSConfig,
	); err != nil {
		return err
	}

	return s.core.Run()
}
