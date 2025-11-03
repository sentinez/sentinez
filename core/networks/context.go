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

package networks

import (
	"context"
	"io"
)

// Context defines the minimal interface representing an HTTP request context.
//
// It abstracts common metadata extracted from an HTTP request, providing
// read-only access to request properties such as headers, query parameters,
// body content, client information, and request metadata.
//
// This interface is typically implemented by a concrete request context
// that represents a single HTTP transaction, used by middleware, handlers,
// or network filters.
type Context interface {
	// Context returns the standard Go context for managing deadlines,
	// cancellations, and request-scoped values.
	Context() context.Context

	// Header returns all request headers as a key-value map.
	Header() map[string]string

	// Queries returns the list of query parameters included in the URL.
	Queries() map[string][]string

	// Path returns the raw path component of the
	// request URI (e.g. /api/v1/users).
	Path() string

	// URI returns the full request URI including
	// scheme, host, path, and query string.
	URI() string

	// Body returns the raw request body as a byte slice.
	Body() []byte

	// ClientIP returns the originating IP address
	// of the client making the request.
	ClientIP() string

	// JA4 returns the JA4 fingerprint string for
	// identifying TLS client characteristics.
	JA4() string

	// TLS indicates whether the request was made
	// over a secure (HTTPS) connection.
	TLS() bool

	// Method returns the HTTP method (e.g. GET, POST, PUT, DELETE).
	Method() string

	// Host returns the request host name (e.g. example.com).
	Host() string

	// StatusCode returns the HTTP response
	// status code associated with this request.
	StatusCode() int

	// Protocol returns the protocol version
	// used for the request (e.g. HTTP/1.1, HTTP/2).
	Protocol() string

	// RemoteAddress returns the full remote
	// address of the client, including port if available.
	RemoteAddress() string
}

// XContext extends Context with mutable capabilities for advanced request and
// response manipulation.
//
// It allows middleware or handlers to modify response properties such as body
// and status code, as well as inspect and manipulate raw headers and streams.
//
// Typically used by reverse proxies, WAF engines, or low-level HTTP frameworks
// that need fine-grained control over request/response processing.
type XContext interface {
	Context

	// SetBody replaces the current response body with the provided byte slice.
	SetBody(body []byte)

	// SetStatusCode sets the HTTP response status code.
	SetStatusCode(code int)

	// ResetResponse resets the current
	// response, discarding any body or headers.
	ResetResponse()

	// VisitReqHeaders iterates over all request headers.
	// The visitor callback receives the header key and value as byte slices.
	VisitReqHeaders(visitor func(k, v []byte))

	// VisitRespHeaders iterates over all response headers.
	VisitRespHeaders(visitor func(k, v []byte))

	// GetReqHeader retrieves the value of a specific request header by key.
	GetReqHeader(k string) string

	// GetRespHeader retrieves the value of a specific response header by key.
	GetRespHeader(k string) string

	// RequestBodyStream returns an io.Reader
	// for streaming the raw request body.
	RequestBodyStream() io.Reader

	// Copy copies data from the given io.Reader into the response body.
	// Useful for proxying large payloads or streaming data directly.
	Copy(src io.Reader) error
}
