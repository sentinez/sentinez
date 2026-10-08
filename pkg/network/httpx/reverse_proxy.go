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

package httpx

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/zlog"
)

func NewReverseProxy(
	target string, trans http.RoundTripper) (corehttp.ReverseProxy, error) {
	urlParsed, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	client := httputil.NewSingleHostReverseProxy(urlParsed)
	client.Transport = trans

	return &ReverseProxy{
		client: client,
		host:   bytesconv.S2b(urlParsed.Host),
	}, nil
}

type ReverseProxy struct {
	client *httputil.ReverseProxy
	host   []byte
}

func (p *ReverseProxy) Serve(ctx corehttp.Context) {
	if p == nil {
		_ = corehttp.NotFound(ctx)
		zlog.Errorf("target not found in reverse proxy memory")
		return
	}

	nctx, ok := ctx.Unwrap().(*Context)
	if !ok {
		_ = corehttp.InternalServerError(ctx)
		zlog.Fatal("request context not supported")
		return
	}

	nctx.SetHost(p.host)

	p.client.ServeHTTP(&recordWriter{
		ResponseWriter: nctx.Response(),
		ctx:            nctx,
	}, nctx.Request())
}

// recordWriter mirrors what the upstream sends into the context, so later
// stages (logging, CDN) see the real status code and body.
type recordWriter struct {
	http.ResponseWriter
	ctx *Context
}

func (w *recordWriter) WriteHeader(code int) {
	w.ctx.ectx.Request.Status = int32(code)
	w.ResponseWriter.WriteHeader(code)
}

func (w *recordWriter) Write(b []byte) (int, error) {
	w.ctx.captureBody(b)
	return w.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach Flush and Hijack, which
// httputil.ReverseProxy needs for streaming and protocol upgrades.
func (w *recordWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
