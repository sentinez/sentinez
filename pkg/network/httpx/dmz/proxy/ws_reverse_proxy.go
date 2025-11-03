package proxydmz

import (
	"github.com/hertz-contrib/reverseproxy"
	httpxdmz "github.com/sentinez/sentinez/pkg/network/httpx/dmz"
)

func NewWSReverseProxy() (*WSReverseProxy, error) {
	proxy := &WSReverseProxy{}
	return proxy, nil
}

type WSReverseProxy struct{}

func (p *WSReverseProxy) Serve(ctx *httpxdmz.Context, target string) {

	uri := ctx.URI()
	if len(uri) != 0 {
		target += string(uri)
	}

	// TODO: forward custom header of sentine-edge

	wsReverseProxy := reverseproxy.NewWSReverseProxy(target)
	wsReverseProxy.ServeHTTP(ctx.Context(), ctx.Unwrap())
}
