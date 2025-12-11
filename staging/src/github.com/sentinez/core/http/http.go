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

package corehttp

import "context"

type Server interface {
	Shutdown(ctx context.Context) error
	ListenAndServe(addr string) error
	Use(mdw ...func(next RequestHandler) RequestHandler)
	Handle(fn RequestHandler)
	ListenAndServeTLS(addr, certFile, keyFile string) error
}

type ReverseProxy interface {
	Serve(ctx Context)
}

const (
	HeaderServer        = "Server"
	HeaderXRequestId    = "X-Request-Id"
	HeaderContentType   = "Content-Type"
	HeaderUpgrade       = "Upgrade"
	HeaderUserAgent     = "User-Agent"
	HeaderXForwardedFor = "X-Forwarded-For"
	HeaderXRealIP       = "X-Real-IP"
	HeaderCacheControl  = "Cache-Control"

	ValueNotFound            = "Not found"
	ValueInternalServerError = "Internal server error"
	ValueAccessDenied        = "Access denied"
	ValueTextPlain           = "text/plain; charset=utf-8"
	ValueTextHTML            = "text/html; charset=utf-8"
	ValueAppJSON             = "application/json; charset=utf-8"

	SchemeSecure   = "https"
	SchemeInsecure = "http"
)
