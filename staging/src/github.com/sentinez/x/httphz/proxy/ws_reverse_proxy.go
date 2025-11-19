package proxydmz

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/reverseproxy"
	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/shared/zlog"
)

func NewWSReverseProxy() (*WSReverseProxy, error) {
	proxy := &WSReverseProxy{}
	return proxy, nil
}

type WSReverseProxy struct{}

func (p *WSReverseProxy) Serve(ctx corehttp.Context, target string) {

	uri := ctx.URI()
	if len(uri) != 0 {
		target += string(uri)
	}

	// TODO: forward custom header of sentine-edge

	rctx, ok := ctx.Unwrap().(*app.RequestContext)
	if !ok {
		_ = ctx.String(
			http.StatusInternalServerError, "Internal server error")

		zlog.Fatal("request context not supported")

		return
	}

	wsReverseProxy := reverseproxy.NewWSReverseProxy(target)
	wsReverseProxy.ServeHTTP(ctx.Context(), rctx)
}
