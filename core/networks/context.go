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

type Context interface {
	Context() context.Context
	Header() map[string]string
	Queries() []string
	Path() string
	URI() string
	Body() []byte
	ClientIP() string
	JA4() string
	TLS() bool
	Method() string
	Host() string
	StatusCode() int
	GetReqProtocol() string
	RemoteAddress() string
}

type XContext interface {
	Context

	// Set body content
	SetBody(body []byte)

	// Set response status code
	SetStatusCode(code int)

	// Reset or rewrite response
	ResetResponse()

	// Header utilities
	VisitReqHeaders(visitor func(k, v []byte))
	VisitRespHeaders(visitor func(k, v []byte))
	GetReqHeader(k string) string
	GetRespHeader(k string) string

	RequestBodyStream() io.Reader
	Copy(src io.Reader) error
}
