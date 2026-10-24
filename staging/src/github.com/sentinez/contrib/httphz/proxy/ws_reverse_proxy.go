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

package proxyhz

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/reverseproxy"
	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/shared/szlog"
)

func NewWSReverseProxy(target string) (*WSReverseProxy, error) {
	proxy := &WSReverseProxy{
		target: target,
	}
	return proxy, nil
}

type WSReverseProxy struct {
	target string
}

func (p *WSReverseProxy) Serve(ctx szhttp.Context) {

	uri := ctx.URI()
	if len(uri) != 0 {
		p.target += string(uri)
	}

	// TODO: forward custom header of sentine-edge

	rctx, ok := ctx.Unwrap().(*app.RequestContext)
	if !ok {
		_ = ctx.String(
			http.StatusInternalServerError, szbytestr.InternalServerError)

		szlog.Errorf("request context not supported")

		return
	}

	wsReverseProxy := reverseproxy.NewWSReverseProxy(p.target)
	wsReverseProxy.ServeHTTP(ctx.Context(), rctx)
}
