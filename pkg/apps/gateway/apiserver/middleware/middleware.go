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
	"net"
	"net/http"
	"strings"

	"github.com/sentinez/core/szhttp/httpconst"
	"github.com/sentinez/pkg/szcolor"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/sentinez/pkg/pools/ingress"
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
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		h.ServeHTTP(lw, clonedR)

		if lw.statusCode >= 400 {
			szlog.Debugf("%s %v code=%+v body=%+v", szcolor.Blue.Add(r.Method),
				szcolor.Green.Add(r.URL.Path),
				szcolor.Status(lw.statusCode),
				string(body))
		}

		event := ingress.Acquire()

		event.Scheme = r.URL.Scheme
		event.Host = r.Host
		event.Path = r.URL.Path
		event.Method = r.Method
		event.Status = int32(lw.statusCode)
		event.RemoteAddress = ip
		event.Protocol = r.Proto
		event.Queries = r.URL.Query().Encode()
		event.UserAgent = r.UserAgent()
		event.ContentType = r.Header.Get(httpconst.HeaderContentType)

		lw.Logging.Info("allow http request", event, event)
	})
}

// CORS allows Cross Origin Resource Sharing, with credentials, from the
// listed origins only (eg "https://console.example.com"). Any other origin
// gets no CORS header, so the browser keeps its page from reading the
// response. Without origins no cross origin call is allowed.
func CORS(origins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowed[origin] = struct{}{}
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				h.ServeHTTP(w, r)
				return
			}

			// the answer depends on the origin: caches must tell them apart
			w.Header().Add("Vary", "Origin")

			if _, ok := allowed[origin]; !ok {
				h.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions &&
				r.Header.Get("Access-Control-Request-Method") != "" {

				preflightHandler(w, r)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// preflightHandler answers the preflight of an allowed origin: it may use
// the methods "GET", "HEAD", "POST", "PUT", "DELETE".
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	// Headers set after WriteHeader are dropped, so cache the preflight first.
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.WriteHeader(http.StatusNoContent)

	grpclog.Infof("Preflight request for %s", r.URL.Path)
}
