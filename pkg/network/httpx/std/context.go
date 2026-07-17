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
	"github.com/sentinez/core/common/bytestr"
	corehttp "github.com/sentinez/core/http"
	httpconst "github.com/sentinez/core/http/const"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	httppb "github.com/sentinez/sentinez/api/gen/go/sentinez/network/http/v1"
	ssync "github.com/sentinez/shared/sync"
)

var (
	_ corehttp.Context = (*Context)(nil)
)

var (
	ctxPool = ssync.NewPoolCtr(func() *Context {
		return &Context{
			request: &httppb.Request{Status: http.StatusOK},
			x:       &edgepb.Context{},
		}
	})
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
	httpCtx.request.Status = http.StatusOK

	httpCtx.resp = resp

	return httpCtx
}

type Context struct {
	req         *http.Request
	resp        http.ResponseWriter
	respBodyBuf bytes.Buffer

	request *httppb.Request
	x       *edgepb.Context
}

// SetRequestId implements corehttp.Context.
func (c *Context) SetRequestId(id string) {
	c.request.Id = id
	c.req.Header.Set(httpconst.HeaderXRequestId, id)
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
func (c *Context) Header(k []byte) []byte {
	values, ok := c.req.Header[string(k)]
	if !ok || len(values) == 0 {
		return nil
	}

	return []byte(values[0])
}

// Headers implements corehttp.Context.
func (c *Context) Headers() map[string][][]byte {
	headers := make(map[string][][]byte)
	c.VisitRequestHeaders(func(key, value []byte) {
		headers[string(key)] = append(headers[string(key)], value)
	})

	return headers
}

// AddResponseHeader implements corehttp.Context.
func (c *Context) AddResponseHeader(key, value []byte) {
	c.resp.Header().Add(string(key), string(value))
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
func (c *Context) RequestIP() []byte {
	// 1. X-Forwarded-For
	if xff := c.Header(bytestr.HeaderXForwardedFor); len(xff) > 0 {
		ips := strings.Split(string(xff), ",")
		return []byte(strings.TrimSpace(ips[0]))
	}

	// 2. X-Real-IP
	if xrip := c.Header(bytestr.HeaderXRealIP); len(xrip) > 0 {
		return xrip
	}

	// 3. Fallback: RemoteAddr
	host, _, err := net.SplitHostPort(c.req.RemoteAddr)
	if err != nil {
		return []byte(c.req.RemoteAddr)
	}
	return []byte(host)
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
func (c *Context) Host() []byte {
	return []byte(c.req.Host)
}

// JA4 implements corehttp.Context.
func (c *Context) JA4() string {
	return c.request.GetFingerprint()
}

// Protocol implements corehttp.Context.
func (c *Context) Protocol() string {
	return c.req.Proto
}

// Queries implements corehttp.Context.
func (c *Context) Queries() map[string][][]byte {
	values := c.req.URL.Query()
	queries := make(map[string][][]byte, len(values))
	for k, vs := range values {
		byteValues := make([][]byte, len(vs))
		for i, v := range vs {
			byteValues[i] = []byte(v)
		}
		queries[k] = byteValues
	}
	return queries
}

// Query implements corehttp.Context.
func (c *Context) Query(k []byte) []byte {
	return []byte(c.req.URL.Query().Get(string(k)))
}

// RemoteAddr implements corehttp.Context.
func (c *Context) RemoteAddr() []byte {
	return []byte(c.req.RemoteAddr)
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
func (c *Context) ResponseHeader() map[string][][]byte {
	headers := make(map[string][][]byte)
	c.VisitResponseHeaders(func(key, value []byte) {
		headers[string(key)] = append(headers[string(key)], value)
	})

	return headers
}

// SetBody implements corehttp.Context.
func (c *Context) SetBody(b []byte) {
	_, _ = c.respBodyBuf.Write(b)
	_, _ = c.resp.Write(b)
}

// SetRequestIP implements corehttp.Context.
func (c *Context) SetRequestIP(ip []byte) {
	c.request.ClientIp = string(ip)
}

func (c *Context) SetHost(host []byte) {
	c.req.Host = string(host)

	if c.req.URL != nil {
		c.req.URL.Host = string(host)
	}

	c.request.Host = string(host)
}

// SetJA4 implements corehttp.Context.
func (c *Context) SetJA4(fingerprint string) {
	c.request.Fingerprint = fingerprint
}

// SetMethod implements corehttp.Context.
func (c *Context) SetMethod(method []byte) {
	c.req.Method = string(method)
}

func (c *Context) SetPath(path []byte) {
	c.req.URL.Path = string(path)
	c.req.URL.RawPath = string(path)

	c.request.Path = path
}

// SetProtocol implements corehttp.Context.
func (c *Context) SetProtocol(p string) {
	c.req.Proto = p
}

// SetQuery implements corehttp.Context.
func (c *Context) SetQuery(k []byte, v ...[]byte) {
	q := c.req.URL.Query()

	q.Del(string(k))

	for _, value := range v {
		q.Add(string(k), string(value))
	}

	c.req.URL.RawQuery = q.Encode()

	// sync protobuf
	var current *httppb.RequestQuery

	for _, query := range c.request.GetQueries() {
		if bytes.Equal(query.Key, []byte(k)) {
			current = query
			break
		}
	}

	if current == nil {
		current = &httppb.RequestQuery{
			Key: []byte(k),
		}
		c.request.Queries = append(c.request.Queries, current)
	}

	current.Values = current.Values[:0]

	for _, value := range v {
		current.Values = append(current.Values, []byte(value))
	}
}

// SetRemoteAddr implements corehttp.Context.
func (c *Context) SetRemoteAddr(addr []byte) {
	c.req.RemoteAddr = string(addr)
}

// SetResponseHeader implements corehttp.Context.
func (c *Context) SetResponseHeader(key []byte, value []byte) {
	c.resp.Header().Set(string(key), string(value))
}

// SetStatusCode implements corehttp.Context.
func (c *Context) SetStatusCode(code int) {
	c.request.Status = int32(code)
	c.resp.WriteHeader(code)
}

func (c *Context) SetURI(u []byte) {
	pURL, err := url.Parse(string(u))
	if err != nil {
		return
	}

	c.req.URL = pURL
	c.request.Uri = []byte(pURL.String())
}

func (c *Context) SetHeader(k, v []byte) {
	c.req.Header.Set(string(k), string(v))

	var current *httppb.RequestHeader
	for _, h := range c.request.Headers {
		if bytes.Equal(h.Key, []byte(k)) {
			current = h
			break
		}
	}

	if current == nil {
		current = &httppb.RequestHeader{
			Key: []byte(k),
		}
		c.request.Headers = append(c.request.Headers, current)
	}

	current.Values = current.Values[:0]
	current.Values = append(current.Values, []byte(v))
}

// StatusCode implements corehttp.Context.
func (c *Context) StatusCode() int {
	return int(c.request.Status)
}

// TLS implements corehttp.Context.
func (c *Context) TLS() bool {
	return c.Scheme() == httpconst.SchemeSecure
}

// URI implements corehttp.Context.
func (c *Context) URI() []byte {
	return []byte(c.req.URL.String())
}

// Unwrap implements corehttp.Context.
func (c *Context) Unwrap() any {
	return c
}

func (c *Context) QueryStr() []byte {
	return []byte(c.req.URL.RawQuery)
}

func (c *Context) Context() context.Context {
	return c.req.Context()
}

func (c *Context) RequestTime() time.Time {
	return c.request.GetTimestamp().AsTime()
}

func (c *Context) Method() []byte {
	return []byte(c.req.Method)
}

func (c *Context) Path() []byte {
	return []byte(c.req.URL.Path)
}

func (c *Context) RequestId() string {
	return c.request.GetId()
}

func (c *Context) JSON(statusCode int, body []byte) error {
	c.SetResponseHeader(bytestr.HeaderContentType, bytestr.ValueAppJSON)
	c.SetStatusCode(statusCode)

	// Use a JSON encoder to write the data
	encoder := json.NewEncoder(c.resp)
	return encoder.Encode(body)
}

func (c *Context) String(statusCode int, msg []byte) error {
	c.SetResponseHeader(bytestr.HeaderContentType, bytestr.ValueTextPlain)
	c.SetStatusCode(statusCode)

	_, err := c.resp.Write(msg)

	return err
}

func (c *Context) Render(statusCode int, component templ.Component) error {
	var buf bytes.Buffer
	if err := component.Render(c.Context(), &buf); err != nil {
		return err
	}

	c.SetResponseHeader(bytestr.HeaderContentType, bytestr.ValueTextHTML)
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
	scheme := httpconst.SchemeInsecure
	if c.req.TLS != nil {
		scheme = httpconst.SchemeSecure
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
	c.respBodyBuf.Reset()

	c.request.Reset()
	c.x.Reset()

	ctxPool.Put(c)
}
