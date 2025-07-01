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

package httpxv2

import (
	"net/http"

	"github.com/valyala/fasthttp"
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
	ctx.SetBodyString("Access denied")
	ctx.Response.Header.Set("Content-Type", "text/plain; charset=utf-8")
}
