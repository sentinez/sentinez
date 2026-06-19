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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/websocket"
	"github.com/sentinez/core"
	corehttp "github.com/sentinez/core/http"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/unsafe"
	"github.com/sentinez/shared/zlog"
)

var (
	_ corehttp.Context = (*Context)(nil)
)

var (
	ctxPool = ssync.NewPool[Context]()
	xPool   = ssync.NewPool[edgepb.Context]()
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		_ = r       // Ignore the request for origin check
		return true // Allow all origins for simplicity
	},
}

func GetContext() *Context {
	return ctxPool.Get()
}

func NewContext(req *http.Request, resp http.ResponseWriter) *Context {
	httpCtx := ctxPool.Get()

	httpCtx.req = req
	httpCtx.resp = resp
	httpCtx.respStatus = 200

	httpCtx.x = xPool.Get()

	return httpCtx
}

type Context struct {
	id          string
	req         *http.Request
	reqTime     time.Time
	resp        http.ResponseWriter
	respStatus  int
	respBodyBuf bytes.Buffer

	x *edgepb.Context
}

// SetRequestId implements corehttp.Context.
func (c *Context) SetRequestId(id string) {
	c.id = id
}

// Extra implements corehttp.Context.
func (c *Context) Extra() *edgepb.Context {
	return c.x
}

// SetExtra implements corehttp.Context.
func (c *Context) SetExtra(x *edgepb.Context) {
	c.x = x
}

// Copy implements corehttp.Context.
func (c *Context) Copy(src io.Reader) error {
	_, err := io.Copy(c.resp, src)
	return err
}

// RequestBodyStream implements corehttp.Context.
func (c *Context) RequestBodyStream() io.Reader {
	return c.req.Body
}

// VisitRequestHeaders implements corehttp.Context.
func (c *Context) VisitRequestHeaders(visitor func(k []byte, v []byte)) {
	for key, values := range c.req.Header {
		for _, v := range values {
			visitor([]byte(key), []byte(v))
		}
	}
}

// VisitResponseHeaders implements corehttp.Context.
func (c *Context) VisitResponseHeaders(visitor func(k []byte, v []byte)) {
	for key, values := range c.resp.Header() {
		for _, v := range values {
			visitor([]byte(key), []byte(v))
		}
	}
}

// Header implements corehttp.Context.
func (c *Context) Header(k string) string {
	values, ok := c.req.Header[k]
	if !ok || len(values) == 0 {
		return ""
	}

	return values[0]
}

// Headers implements corehttp.Context.
func (c *Context) Headers() map[string]string {
	headers := make(map[string]string)
	c.VisitRequestHeaders(func(key, value []byte) {
		headers[(string(key))] = string(value)
	})

	return headers
}

// AddResponseHeader implements corehttp.Context.
func (c *Context) AddResponseHeader(key string, value string) {
	c.resp.Header().Add(key, value)
}

// Body implements corehttp.Context.
func (c *Context) Body() []byte {
	body, err := io.ReadAll(c.req.Body)
	if err != nil {
		return nil
	}

	c.req.Body = io.NopCloser(bytes.NewBuffer(body))

	return body
}

// RequestIP implements corehttp.Context.
func (c *Context) RequestIP() string {
	// 1. X-Forwarded-For
	if xff := c.Header(corehttp.HeaderXForwardedFor); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 2. X-Real-IP
	if xrip := c.req.Header.Get(corehttp.HeaderXRealIP); xrip != "" {
		return xrip
	}

	// 3. Fallback: RemoteAddr
	host, _, err := net.SplitHostPort(c.req.RemoteAddr)
	if err != nil {
		return c.req.RemoteAddr
	}
	return host
}

// File implements corehttp.Context.
func (c *Context) File(path string) error {
	http.ServeFile(c.resp, c.req, path)
	return nil
}

// Flush implements corehttp.Context.
func (c *Context) Flush() error {
	flusher, ok := c.resp.(http.Flusher)
	if !ok {
		return fmt.Errorf("response writer does not support flush")
	}

	flusher.Flush()
	return nil
}

// Host implements corehttp.Context.
func (c *Context) Host() string {
	return c.req.Host
}

// JA4 implements corehttp.Context.
func (c *Context) JA4() string {
	return ""
}

// Protocol implements corehttp.Context.
func (c *Context) Protocol() string {
	return c.req.Proto
}

// Queries implements corehttp.Context.
func (c *Context) Queries() map[string][]string {
	return c.req.URL.Query()
}

// Query implements corehttp.Context.
func (c *Context) Query(k string) string {
	return c.req.URL.Query().Get(k)
}

// RemoteAddr implements corehttp.Context.
func (c *Context) RemoteAddr() string {
	return c.req.RemoteAddr
}

// ResetResponse implements corehttp.Context.
func (c *Context) ResetResponse() {
	c.respBodyBuf.Reset()
}

// ResponseBody implements corehttp.Context.
func (c *Context) ResponseBody() []byte {
	return c.respBodyBuf.Bytes()
}

// ResponseHeader implements corehttp.Context.
func (c *Context) ResponseHeader() map[string]string {
	headers := make(map[string]string)
	c.VisitResponseHeaders(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	return headers
}

// SetBody implements corehttp.Context.
func (c *Context) SetBody(b []byte) {
	_, _ = c.respBodyBuf.Write(b)
	_, _ = c.resp.Write(b)
}

// SetRequestIP implements corehttp.Context.
func (c *Context) SetRequestIP(_ string) {
	zlog.Fatal("[stdhttp] unimplemented")
}

// SetHost implements corehttp.Context.
func (c *Context) SetHost(h string) {
	c.req.Host = h
}

// SetJA4 implements corehttp.Context.
func (c *Context) SetJA4(_ string) {
	zlog.Fatal("[stdhttp] unimplemented")
}

// SetMethod implements corehttp.Context.
func (c *Context) SetMethod(method string) {
	c.req.Method = method
}

// SetPath implements corehttp.Context.
func (c *Context) SetPath(p string) {
	c.req.URL.Path = p
}

// SetProtocol implements corehttp.Context.
func (c *Context) SetProtocol(p string) {
	c.req.Proto = p
}

// SetQuery implements corehttp.Context.
func (c *Context) SetQuery(_ string, _ ...string) {
	zlog.Fatal("[stdhttp] unimplemented")
}

// SetRemoteAddr implements corehttp.Context.
func (c *Context) SetRemoteAddr(addr string) {
	c.req.RemoteAddr = addr
}

// SetResponseHeader implements corehttp.Context.
func (c *Context) SetResponseHeader(key string, value string) {
	c.resp.Header().Set(key, value)
}

// SetStatusCode implements corehttp.Context.
func (c *Context) SetStatusCode(code int) {
	c.respStatus = code
	c.resp.WriteHeader(code)
}

// SetURI implements corehttp.Context.
func (c *Context) SetURI(u string) {
	pURL, err := url.Parse(u)
	if err != nil {
		c.req.URL = pURL
	}
}

// StatusCode implements corehttp.Context.
func (c *Context) StatusCode() int {
	return c.respStatus
}

// TLS implements corehttp.Context.
func (c *Context) TLS() bool {
	return c.Scheme() == corehttp.SchemeSecure
}

// URI implements corehttp.Context.
func (c *Context) URI() string {
	return c.req.URL.String()
}

// Unwrap implements corehttp.Context.
func (c *Context) Unwrap() any {
	return c
}

func (c *Context) QueryStr() string {
	return c.req.URL.RawQuery
}

func (c *Context) SetHeader(k, v string) {
	c.req.Header.Set(k, v)
}

func (c *Context) Context() context.Context {
	return c.req.Context()
}

func (c *Context) RequestTime() time.Time {
	return c.reqTime
}

func (c *Context) Method() string {
	return c.req.Method
}

func (c *Context) Path() string {
	return c.req.URL.Path
}

func (c *Context) RequestId() string {
	return c.id
}

func (c *Context) JSON(statusCode int, body []byte) error {
	c.SetResponseHeader(corehttp.HeaderContentType, corehttp.ValueAppJSON)
	c.SetResponseHeader(corehttp.HeaderServer, core.Name)
	c.SetStatusCode(statusCode)

	// Use a JSON encoder to write the data
	encoder := json.NewEncoder(c.resp)
	return encoder.Encode(body)
}

func (c *Context) String(statusCode int, msg string) error {
	c.SetResponseHeader(corehttp.HeaderContentType, corehttp.ValueTextPlain)
	c.SetResponseHeader(corehttp.HeaderServer, core.Name)
	c.SetStatusCode(statusCode)

	_, err := c.resp.Write(unsafe.S2B(msg))

	return err
}

func (c *Context) Render(statusCode int, component templ.Component) error {
	var buf bytes.Buffer
	if err := component.Render(c.Context(), &buf); err != nil {
		return err
	}

	c.SetResponseHeader(corehttp.HeaderContentType, corehttp.ValueTextHTML)
	c.SetResponseHeader(corehttp.HeaderServer, core.Name)
	c.SetStatusCode(statusCode)

	_, err := c.resp.Write(buf.Bytes())
	return err
	//return component.Render(c.Context(), c.resp)
}

func (c *Context) Upgrade() (*websocket.Conn, error) {
	conn, err := upgrade.Upgrade(c.resp, c.req, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Context) Scheme() string {
	scheme := corehttp.SchemeInsecure
	if c.req.TLS != nil {
		scheme = corehttp.SchemeSecure
	}

	return scheme
}

func (c *Context) Request() *http.Request {
	return c.req
}

func (c *Context) Response() http.ResponseWriter {
	return c.resp
}

func Release(c *Context) {
	c.req = nil

	c.resp = nil
	c.respStatus = http.StatusOK
	c.respBodyBuf.Reset()

	c.x.Reset()
	xPool.Put(c.x)

	ctxPool.Put(c)
}
