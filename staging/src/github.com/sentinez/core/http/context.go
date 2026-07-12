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

import (
	"context"
	"io"
	"time"

	"github.com/a-h/templ"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
)

type RequestContext interface {
	Context() context.Context

	Headers() map[string][][]byte
	Queries() map[string][][]byte

	Query(k []byte) []byte
	QueryStr() []byte

	Header(k []byte) []byte

	Path() []byte
	URI() []byte

	Body() []byte
	JA4() string
	TLS() bool

	Method() []byte
	Host() []byte

	StatusCode() int
	Protocol() string
	Scheme() string

	RemoteAddr() []byte
	RequestIP() []byte

	RequestId() string
	RequestTime() time.Time

	SetHeader(key, value []byte)
	SetQuery(key []byte, values ...[]byte)

	SetPath(p []byte)
	SetURI(u []byte)

	SetBody(b []byte)
	SetRequestId(id string)

	SetRequestIP(ip []byte)

	SetJA4(fingerprint string)

	SetMethod(m []byte)
	SetHost(h []byte)

	SetStatusCode(code int)
	SetProtocol(p string)

	SetRemoteAddr(addr []byte)
}

type ResponseWriter interface {
	ResponseHeader() map[string][][]byte
	ResponseBody() []byte

	SetResponseHeader(key, value []byte)
	AddResponseHeader(key, value []byte)
	ResetResponse()

	String(statusCode int, msg []byte) error
	JSON(statusCode int, body []byte) error
	File(path string) error
	Flush() error
	Copy(src io.Reader) error
}

type Context interface {
	RequestContext
	ResponseWriter
	Unwrap() any

	Extra() *edgepb.Context
	SetExtra(x *edgepb.Context)

	Render(statusCode int, component templ.Component) error
	RequestBodyStream() io.Reader
	VisitRequestHeaders(visitor func(k, v []byte))
	VisitResponseHeaders(visitor func(k, v []byte))
}

type RequestHandler func(Context) error
