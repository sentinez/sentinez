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

package corers

import (
	"net/http"

	"github.com/corazawaf/coraza/v3/types"
	corehttp "github.com/sentinez/core/http"
)

// interceptor for fasthttp
// ref: https://github.com/corazawaf/coraza/tree/main/http
type interceptor struct {
	tx          types.Transaction
	wroteHeader bool
	proto       string
}

func (i *interceptor) writeResponseHeader(ctx corehttp.Context) {
	if i.wroteHeader {
		return
	}

	ctx.VisitResponseHeaders(func(k, v []byte) {
		i.tx.AddResponseHeader(string(k), string(v))
	})

	if it := i.tx.ProcessResponseHeaders(ctx.StatusCode(), i.proto); it != nil {
		return
	}

	i.wroteHeader = true
}

func (i *interceptor) writeResponseBody(
	ctx corehttp.Context) (*types.Interruption, error) {
	if i.tx.IsInterrupted() {
		return nil, nil
	}

	if !i.wroteHeader {
		i.writeResponseHeader(ctx)
	}

	if i.tx.IsResponseBodyAccessible() && i.tx.IsResponseBodyProcessable() {
		it, _, err := i.tx.WriteResponseBody(ctx.Body())
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
