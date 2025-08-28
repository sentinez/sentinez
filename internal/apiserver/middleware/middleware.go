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

// Package middleware provide http handler - net/http
package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	httppb "github.com/sentinez/sentinez/api/gen/go/sentinez/std/net/http/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"google.golang.org/grpc/grpclog"
)

// Logging logs the request body when the response status code is not 200
// This addresses the issue of being unable to retrieve the request body in the
// customErrorHandler middleware.
// nolint:funlen
func Logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := newLogResponseWriter(w)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w,
				fmt.Sprintf("grpc server read request body err %+v", err),
				http.StatusBadRequest)

			return
		}

		clonedR := r.Clone(r.Context())
		clonedR.Body = io.NopCloser(bytes.NewReader(body))

		h.ServeHTTP(lw, clonedR)

		if lw.statusCode >= 400 {
			zlog.Debugf("%s %v code=%+v body=%+v", color.Blue.Add(r.Method),
				color.Green.Add(r.URL.Path),
				color.Status(lw.statusCode),
				string(body))
		}

		lw.Logger.Info("allow http request", &httppb.Log4HTTP{
			ReqScheme:     r.URL.Scheme,
			ReqHost:       r.Host,
			ReqPath:       r.URL.Path,
			ReqMethod:     r.Method,
			RespStatus:    int32(lw.statusCode),
			ReqRemoteAddr: r.RemoteAddr,
			ReqProtocol:   r.Proto,
			ReqQuery:      r.URL.RawQuery,
			UserAgent:     r.UserAgent(),
			ContentType:   r.Header.Get("Content-Type"),
		})
	})
}

// AllowCORS allows Cross Origin Resource Sharing from any origin.
// Don't do this without consideration in production systems.
func AllowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)

			// Fix cache issue
			w.Header().Set("Vary", "Origin")

			// If cookies need to be sent
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" &&
				r.Header.Get("Access-Control-Request-Method") != "" {

				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

// preflightHandler adds the necessary headers in order to serve
// CORS from any origin using the methods "GET", "HEAD", "POST", "PUT", "DELETE"
// We insist, don't do this without consideration in production systems.
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))

	// Fix: Add status code to avoid error
	w.WriteHeader(http.StatusNoContent)

	// Fix: Add Max-Age to optimize
	w.Header().Set("Access-Control-Max-Age", "86400")

	grpclog.Infof("Preflight request for %s", r.URL.Path)
}
