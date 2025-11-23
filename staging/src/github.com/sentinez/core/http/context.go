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
	"github.com/gorilla/websocket"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
)

type RequestContext interface {
	Context() context.Context
	Headers() map[string]string
	Queries() map[string][]string
	Query(k string) string
	QueryStr() string
	Header(k string) string
	Path() string
	URI() string
	Body() []byte
	JA4() string
	TLS() bool
	Method() string
	Host() string
	StatusCode() int
	Protocol() string
	Scheme() string
	RemoteAddr() string
	RequestIP() string
	RequestId() string
	RequestTime() time.Time

	SetHeader(key, value string)
	SetQuery(key string, values ...string)
	SetPath(p string)
	SetURI(u string)
	SetBody(b []byte)
	SetRequestId(id string)
	SetRequestIP(ip string)
	SetJA4(fingerprint string)
	SetMethod(m string)
	SetHost(h string)
	SetStatusCode(code int)
	SetProtocol(p string)
	SetRemoteAddr(addr string)
}

type ResponseWriter interface {
	ResponseHeader() map[string]string
	ResponseBody() []byte

	SetResponseHeader(key, value string)
	AddResponseHeader(key, value string)
	ResetResponse()

	String(statusCode int, msg string) error
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

	Upgrade() (*websocket.Conn, error)
	Render(statusCode int, component templ.Component) error
	RequestBodyStream() io.Reader
	VisitRequestHeaders(visitor func(k, v []byte))
	VisitResponseHeaders(visitor func(k, v []byte))
}

type RequestHandler func(Context) error
