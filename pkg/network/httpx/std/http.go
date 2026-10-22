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

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/sentinez/shared/szlog"
)

func HandlerFunc(path string, handler szhttp.RequestHandler) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := httpx.NewContext(r, w)

		_ = handler(ctx)

		_ = ctx.Close()
	})
}

func ListenAndServe(addr string, opts ...szhttp.ServerOption) error {
	option := szhttp.NewOption(opts...)

	if option.CertFile != "" && option.CertKeyFile != "" {
		if option.TLSConfig != nil {
			szlog.Warnf("network: http ignore TLS config")
		}

		return http.ListenAndServeTLS(addr,
			option.CertFile, option.CertKeyFile, nil)
	}

	return http.ListenAndServe(addr, nil)
}

func Shutdown() error {
	// Note: http.Server does not have a Shutdown method in the standard library
	// This is a placeholder for future implementation if needed.
	return nil
}

func Convert(handler szhttp.RequestHandler) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		StandardConverter(handler, resp, req)
	}
}

func StandardConverter(handler szhttp.RequestHandler,
	resp http.ResponseWriter, req *http.Request) {

	ctx := httpx.NewContext(req, resp)

	if err := handler(ctx); err != nil {
		http.Error(resp, err.Error(), ctx.StatusCode())
	}

	_ = ctx.Close()
}
