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

package httpxf1

import (
	"net/http"

	"github.com/sentinez/sentinez/pkg/templ"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func acquireRequest() *fasthttp.Request {
	return fasthttp.AcquireRequest()
}

func acquireResponse() *fasthttp.Response {
	return fasthttp.AcquireResponse()
}

// Do perform an HTTP request using the provided context and URI.
func Do(ctx Context, uri string) error {
	req := acquireRequest()
	resp := acquireResponse()

	ctx.Request.CopyTo(req)
	req.SetRequestURI(uri)

	if err := fasthttp.Do(req, resp); err != nil {
		return err
	}

	ctx.SetStatusCode(resp.StatusCode())

	resp.Header.All()(func(k, v []byte) bool {
		ctx.Response.Header.SetBytesKV(k, v)
		return true
	})

	ctx.Response.SetBodyRaw(resp.Body())

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)

	return nil
}

func Forbidden(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(http.StatusForbidden)
	ctx.Response.Header.Set("Content-Type", "text/plain; charset=utf-8")

	ctx.SetContentType("text/html; charset=utf-8")
	rCtx := convertRequestContext(ctx)
	err := templ.Forbidden().Render(rCtx.Context(), ctx.Response.BodyWriter())
	if err != nil {
		ctx.SetBodyString("Access denied")
	}
}

func convertRequestContext(ctx *fasthttp.RequestCtx) *http.Request {
	r := new(http.Request)

	if err := fasthttpadaptor.ConvertRequest(ctx, r, true); err != nil {
		ctx.Error("failed to convert request context",
			fasthttp.StatusInternalServerError)
		return nil
	}

	return r
}

func wrapHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		identifier(ctx)
		next(ctx)
	}
}
