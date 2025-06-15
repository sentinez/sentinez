// Copyright 2025 Sentinez Labs.
// Copyright 2022 Juan Pablo Tosso and the OWASP Coraza contributors
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

package http2sec

import (
	"net/http"

	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/valyala/fasthttp"
)

// interceptor for fasthttp
// ref: https://github.com/corazawaf/coraza/tree/main/http
type interceptor struct {
	ctx                *fasthttp.RequestCtx
	tx                 types.Transaction
	statusCode         int
	isWriteHeaderFlush bool
	wroteHeader        bool
	proto              string
}

func (i *interceptor) flushWriteHeader() {
	if !i.isWriteHeaderFlush {
		i.ctx.SetStatusCode(i.statusCode)
		i.isWriteHeaderFlush = true
	}
}

func (i *interceptor) overrideWriteHeader(statusCode int) {
	i.statusCode = statusCode
}

func (i *interceptor) cleanHeaders() {
	i.ctx.Response.Header.Reset()
}

func (i *interceptor) WriteHeader(statusCode int) {
	if i.wroteHeader {
		zlog.Debug("httpx.secure.http2: skip writing header")
		return
	}

	i.ctx.Response.Header.VisitAll(func(k, v []byte) {
		i.tx.AddResponseHeader(string(k), string(v))
	})

	i.statusCode = statusCode
	if it := i.tx.ProcessResponseHeaders(statusCode, i.proto); it != nil {
		i.cleanHeaders()
		i.ctx.Request.Header.Set("Content-Length", "0")
		i.overrideWriteHeader(obtainStatusCodeFromInterruptionOrDefault(
			it,
			i.statusCode,
		))
		i.flushWriteHeader()

		return
	}

	i.wroteHeader = true
}

func (i *interceptor) Write(b []byte) (int, error) {
	if i.tx.IsInterrupted() {
		return len(b), nil
	}

	if !i.wroteHeader {
		i.WriteHeader(http.StatusOK)
	}

	if i.tx.IsResponseBodyAccessible() && i.tx.IsResponseBodyProcessable() {
		it, n, err := i.tx.WriteResponseBody(b)
		if it != nil {
			i.cleanHeaders()
			i.ctx.Response.Header.Set("Content-Length", "0")
			i.overrideWriteHeader(obtainStatusCodeFromInterruptionOrDefault(
				it,
				i.statusCode,
			))
			i.flushWriteHeader()

			return len(b), nil
		}

		return n, err
	}

	i.flushWriteHeader()

	return i.ctx.Write(b)
}

func obtainStatusCodeFromInterruptionOrDefault(
	it *types.Interruption, defaultStatusCode int) int {

	if it.Action == "deny" {
		statusCode := it.Status
		if statusCode == 0 {
			statusCode = 403
		}

		return statusCode
	}
	return defaultStatusCode
}
