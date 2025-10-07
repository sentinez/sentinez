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

package httpfsec

import (
	"net/http"

	"github.com/corazawaf/coraza/v3/types"
	httpxf1 "github.com/sentinez/sentinez/pkg/core/net/httpx/f1"
	"github.com/sentinez/sentinez/pkg/stdcmn/zlog"
)

// interceptor for fasthttp
// ref: https://github.com/corazawaf/coraza/tree/main/http
type interceptor struct {
	tx          types.Transaction
	wroteHeader bool
	proto       string
}

func (i *interceptor) WriteResponseHeader(ctx *httpxf1.Context) {
	if i.wroteHeader {
		zlog.Debug("httpx.secure.http2: skip writing header")
		return
	}

	ctx.Response.Header.All()(func(k, v []byte) bool {
		i.tx.AddResponseHeader(string(k), string(v))
		return true
	})

	if it := i.tx.ProcessResponseHeaders(
		ctx.Response.StatusCode(), i.proto); it != nil {
		return
	}

	i.wroteHeader = true
}

func (i *interceptor) WriteResponseBody(
	ctx *httpxf1.Context) (*types.Interruption, error) {
	if i.tx.IsInterrupted() {
		return nil, nil
	}

	if !i.wroteHeader {
		i.WriteResponseHeader(ctx)
	}

	if i.tx.IsResponseBodyAccessible() && i.tx.IsResponseBodyProcessable() {
		zlog.Debug("let's write response body ", len(ctx.Response.Body()))
		it, _, err := i.tx.WriteResponseBody(ctx.Response.Body())
		if err != nil {

			return nil, err
		}

		if it != nil {

			return it, nil
		}

		return i.tx.ProcessResponseBody()
	}

	return nil, nil
}

func obtainStatusCodeFromInterruptionOrDefault(
	it *types.Interruption, defaultStatusCode int) int {

	if it == nil {
		return defaultStatusCode
	}

	if it.Status != 0 {
		return it.Status
	}

	if it.Action == "deny" {
		return http.StatusForbidden
	}

	return defaultStatusCode
}
