package proxy

import (
	"github.com/hertz-contrib/reverseproxy"
	httpxhz "github.com/sentinez/sentinez/pkg/network/httpx/hz"
)

func NewWSReverseProxy() (*WSReverseProxy, error) {
	proxy := &WSReverseProxy{}
	return proxy, nil
}

type WSReverseProxy struct{}

func (p *WSReverseProxy) Serve(ctx *httpxhz.Context, target string) {

	uri := ctx.Request.URI().RequestURI()
	if len(uri) != 0 {
		target += string(uri)
	}

	// TODO: forward custom header of sentine-egde

	wsReverseProxy := reverseproxy.NewWSReverseProxy(target)
	wsReverseProxy.ServeHTTP(ctx.Context(), ctx.RequestContext)
}
