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

package stdhttpx

import (
	"net/http"

	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/sentinez"
	sids "github.com/sentinez/shared/ids"
)

func HandlerFunc(path string, handler corehttp.RequestHandler) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r, w)

		_ = handler(ctx)

		ctx.Release()
	})
}

func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func Shutdown() error {
	// Note: http.Server does not have a Shutdown method in the standard library
	// This is a placeholder for future implementation if needed.
	return nil
}

func Convert(handler corehttp.RequestHandler) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		rctx := NewContext(req, resp)

		requestId := sids.NewNanoID(sentinez.PrefixRequestID)
		rctx.req.Header.Set(corehttp.HeaderXRequest, requestId)

		rctx.ctx = corehttp.SetRequestTime(rctx.ctx)

		if err := handler(rctx); err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
		}

		rctx.Release()
	}
}
