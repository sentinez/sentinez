// Copyright 2011 The Go Authors.
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

package stdproxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	corehttp "github.com/sentinez/sentinez/core/http"
	"github.com/sentinez/sentinez/pkg/common/syncx"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func NewReverseProxy() (*ReverseProxy, error) {
	return &ReverseProxy{
		pool: syncx.NewPool[httputil.ReverseProxy](),
	}, nil
}

type ReverseProxy struct {
	pool *syncx.Pool[httputil.ReverseProxy]
}

func (p *ReverseProxy) Serve(ctx corehttp.Context, target string) {
	url, err := url.Parse(target)
	if err != nil {
		_ = httpxcmn.InternalServerError(ctx)
		return
	}

	rproxy := p.pool.Get()

	rproxy.Director = func(req *http.Request) {
		rewriteRequestURL(req, url)
		req.Host = url.Host
	}

	if url.Scheme == corehttp.SchemeInsecure {
		rproxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	nctx, ok := ctx.Unwrap().(*stdhttpx.Context)
	if !ok {
		_ = httpxcmn.InternalServerError(ctx)
		zlog.Fatal("request context not supported")
		return
	}

	rproxy.ServeHTTP(nctx.Response(), nctx.Request())
}

func rewriteRequestURL(req *http.Request, target *url.URL) {
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
