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
	corecontext "github.com/sentinez/core/context"
	corehttp "github.com/sentinez/core/http"
	httpconst "github.com/sentinez/core/http/const"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/network/http/v1"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/store/ja4"
	"github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ corehttp.Context = (*Context)(nil)
	_ io.Closer        = (*Context)(nil)
)

var (
	ctxPool = sync.NewPoolCtr(func() *Context {
		return &Context{
			ectx: &edgepb.Context{
				X:       &edgepb.ContextExtra{},
				Request: &httppb.Request{},
			},
		}
	})
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		_ = r       // Ignore the request for origin check
		return true // Allow all origins for simplicity
	},
}

func NewContext(req *http.Request, resp http.ResponseWriter) *Context {
	zlog.Debug("http context: connection: new context")
	httpCtx := ctxPool.Get()

	httpCtx.req = req
	httpCtx.resp = resp

	transport := corecontext.GetTransport(req.Context())
	if transport != nil {
		connId := transport.GetConnId()
		zlog.Debugf("context: found fingerprint: %s", ja4.Get(connId))
		httpCtx.ectx.Request.Fingerprint = ja4.Get(connId)
	}

	httpCtx.ectx.Transport = transport

	httpCtx.ectx.Request.Status = http.StatusOK
	httpCtx.ectx.Request.Timestamp = timestamppb.Now()

	return httpCtx
}

type Context struct {
	req         *http.Request
	resp        http.ResponseWriter
	respBodyBuf bytes.Buffer
	ectx        *edgepb.Context
}

// X implements [corehttp.Context].
func (c *Context) X() *edgepb.ContextExtra {
	return c.ectx.GetX()
}

// SetRequestId implements corehttp.Context.
func (c *Context) SetRequestId(id string) {
	c.ectx.Request.Id = id
	c.req.Header.Set(httpconst.HeaderXRequestId, id)
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
	return c.ectx.GetRequest().GetFingerprint()
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

func (c *Context) SetPath(path []byte) {
	c.req.URL.Path = string(path)
	c.req.URL.RawPath = string(path)

	c.ectx.Request.Path = path
}

func (c *Context) SetHost(host []byte) {
	sHost := bytesconv.B2s(host)
	c.req.Host = sHost

	if c.req.URL != nil {
		c.req.URL.Host = sHost
	}

	c.ectx.Request.Host = sHost
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

	for _, query := range c.ectx.GetRequest().GetQueries() {
		if bytes.Equal(query.Key, []byte(k)) {
			current = query
			break
		}
	}

	if current == nil {
		current = &httppb.RequestQuery{
			Key: []byte(k),
		}
		c.ectx.Request.Queries = append(c.ectx.Request.Queries, current)
	}

	current.Values = current.Values[:0]

	for _, value := range v {
		current.Values = append(current.Values, []byte(value))
	}
}

// SetResponseHeader implements corehttp.Context.
func (c *Context) SetResponseHeader(key []byte, value []byte) {
	c.resp.Header().Set(string(key), string(value))
}

// SetStatusCode implements corehttp.Context.
func (c *Context) SetStatusCode(code int) {
	c.ectx.Request.Status = int32(code)
	c.resp.WriteHeader(code)
}

func (c *Context) SetURI(u []byte) {
	pURL, err := url.Parse(string(u))
	if err != nil {
		return
	}

	c.req.URL = pURL
	c.ectx.Request.Uri = []byte(pURL.String())
}

func (c *Context) SetHeader(k, v []byte) {
	c.req.Header.Set(string(k), string(v))

	var current *httppb.RequestHeader
	for _, h := range c.ectx.GetRequest().GetHeaders() {
		if bytes.Equal(h.Key, []byte(k)) {
			current = h
			break
		}
	}

	if current == nil {
		current = &httppb.RequestHeader{
			Key: []byte(k),
		}
		c.ectx.Request.Headers = append(c.ectx.Request.Headers, current)
	}

	current.Values = current.Values[:0]
	current.Values = append(current.Values, []byte(v))
}

// StatusCode implements corehttp.Context.
func (c *Context) StatusCode() int {
	return int(c.ectx.GetRequest().GetStatus())
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
	return c.ectx.GetRequest().GetTimestamp().AsTime()
}

func (c *Context) Method() []byte {
	return []byte(c.req.Method)
}

func (c *Context) Path() []byte {
	return []byte(c.req.URL.Path)
}

func (c *Context) RequestId() string {
	return c.ectx.GetRequest().GetId()
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

func (c *Context) Close() error {
	c.req = nil
	c.resp = nil
	c.ectx.Transport = nil
	c.ectx.Metadata = nil

	c.ectx.X.Reset()
	c.ectx.Request.Reset()

	c.respBodyBuf.Reset()

	ctxPool.Put(c)

	return nil
}
