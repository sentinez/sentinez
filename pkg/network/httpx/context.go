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

package httpx

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
	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szcontext"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/szrand"
	"github.com/sentinez/shared/szstore/ja4"
	"github.com/sentinez/shared/szsync"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ szhttp.Context = (*Context)(nil)
	_ io.Closer      = (*Context)(nil)
)

var (
	_ctxPool = szsync.NewPoolCtr(func() *Context {
		return &Context{
			BaseContext: &szhttp.BaseContext{},
			ectx: &edgepb.Context{
				X:       &edgepb.ContextExtra{},
				Request: &httppb.Request{},
			},
		}
	})
)

// _maxCapturedBody bounds how much of a proxied response is kept for
// ResponseBody; larger responses are streamed through but not captured.
const _maxCapturedBody = 4 << 20

// _maxPooledBodyBuf drops oversized buffers instead of pooling them.
const _maxPooledBodyBuf = 64 << 10

var _upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		_ = r       // Ignore the request for origin check
		return true // Allow all origins for simplicity
	},
}

func NewContext(req *http.Request, resp http.ResponseWriter) *Context {
	szlog.Debug("http context: connection: new context")
	hctx := _ctxPool.Get()
	hctx.SetRequestId(szrand.NewXId(core.PrefixXRequestIdBytes))

	hctx.req = req
	hctx.resp = resp

	transport := szcontext.GetTransport(req.Context())
	if transport != nil {
		fingerprint := ja4.Get(transport.GetConnId())
		szlog.Debugf("context: found fingerprint: %s", fingerprint)
		hctx.ectx.Request.Fingerprint = fingerprint
	}

	hctx.ectx.Transport = transport
	hctx.ectx.Request.Status = http.StatusOK
	hctx.ectx.Request.Timestamp = timestamppb.Now()

	return hctx
}

type Context struct {
	*szhttp.BaseContext
	req         *http.Request
	resp        http.ResponseWriter
	respBodyBuf bytes.Buffer
	// bodyOverflow is set once a captured body exceeds _maxCapturedBody.
	bodyOverflow bool
	ectx         *edgepb.Context
}

// X implements [szhttp.Context].
func (c *Context) X() *edgepb.ContextExtra {
	return c.ectx.GetX()
}

// Copy implements szhttp.Context.
func (c *Context) Copy(src io.Reader) error {
	_, err := io.Copy(c.resp, src)
	return err
}

// RequestBodyStream implements szhttp.Context.
func (c *Context) RequestBodyStream() io.Reader {
	return c.req.Body
}

// VisitRequestHeaders implements szhttp.Context.
func (c *Context) VisitRequestHeaders(visitor func(k []byte, v []byte)) {
	for key, values := range c.req.Header {
		for _, v := range values {
			visitor([]byte(key), []byte(v))
		}
	}
}

// VisitResponseHeaders implements szhttp.Context.
func (c *Context) VisitResponseHeaders(visitor func(k []byte, v []byte)) {
	for key, values := range c.resp.Header() {
		for _, v := range values {
			visitor([]byte(key), []byte(v))
		}
	}
}

// Header implements szhttp.Context.
func (c *Context) Header(k []byte) []byte {
	values, ok := c.req.Header[string(k)]
	if !ok || len(values) == 0 {
		return nil
	}

	return []byte(values[0])
}

// Headers implements szhttp.Context.
func (c *Context) Headers() map[string][][]byte {
	headers := make(map[string][][]byte)
	c.VisitRequestHeaders(func(key, value []byte) {
		headers[string(key)] = append(headers[string(key)], value)
	})

	return headers
}

// AddResponseHeader implements szhttp.Context.
func (c *Context) AddResponseHeader(key, value []byte) {
	c.resp.Header().Add(string(key), string(value))
}

// Body implements szhttp.Context.
func (c *Context) Body() []byte {
	body, err := io.ReadAll(c.req.Body)
	if err != nil {
		return nil
	}

	c.req.Body = io.NopCloser(bytes.NewBuffer(body))

	return body
}

// RequestIP implements szhttp.Context.
func (c *Context) RequestIP() []byte {
	// 1. X-Forwarded-For
	if xff := c.Header(szbytestr.HeaderXForwardedFor); len(xff) > 0 {
		ips := strings.Split(string(xff), ",")
		return []byte(strings.TrimSpace(ips[0]))
	}

	// 2. X-Real-IP
	if xrip := c.Header(szbytestr.HeaderXRealIP); len(xrip) > 0 {
		return xrip
	}

	// 3. Fallback: RemoteAddr
	host, _, err := net.SplitHostPort(c.req.RemoteAddr)
	if err != nil {
		return []byte(c.req.RemoteAddr)
	}
	return []byte(host)
}

// File implements szhttp.Context.
func (c *Context) File(path string) error {
	http.ServeFile(c.resp, c.req, path)
	return nil
}

// Flush implements szhttp.Context.
func (c *Context) Flush() error {
	flusher, ok := c.resp.(http.Flusher)
	if !ok {
		return fmt.Errorf("response writer does not support flush")
	}

	flusher.Flush()
	return nil
}

// Host implements szhttp.Context.
func (c *Context) Host() []byte {
	return []byte(c.req.Host)
}

// JA4 implements szhttp.Context.
func (c *Context) JA4() string {
	return c.ectx.GetRequest().GetFingerprint()
}

// Protocol implements szhttp.Context.
func (c *Context) Protocol() string {
	return c.req.Proto
}

// Queries implements szhttp.Context.
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

// Query implements szhttp.Context.
func (c *Context) Query(k []byte) []byte {
	return []byte(c.req.URL.Query().Get(string(k)))
}

// RemoteAddr implements szhttp.Context.
func (c *Context) RemoteAddr() []byte {
	return []byte(c.req.RemoteAddr)
}

// ResetResponse implements szhttp.Context.
func (c *Context) ResetResponse() {
	c.respBodyBuf.Reset()
	c.bodyOverflow = false
}

// ResponseBody implements szhttp.Context. It returns nil when the body
// was too large to capture.
func (c *Context) ResponseBody() []byte {
	if c.bodyOverflow {
		return nil
	}

	return c.respBodyBuf.Bytes()
}

// ResponseHeader implements szhttp.Context.
func (c *Context) ResponseHeader() map[string][][]byte {
	headers := make(map[string][][]byte)
	c.VisitResponseHeaders(func(key, value []byte) {
		headers[string(key)] = append(headers[string(key)], value)
	})

	return headers
}

// SetBody implements szhttp.Context.
func (c *Context) SetBody(b []byte) {
	_, _ = c.respBodyBuf.Write(b)
	_, _ = c.resp.Write(b)
}

func (c *Context) SetPath(path []byte) {
	c.req.URL.Path = string(path)
	c.req.URL.RawPath = string(path)

	// Callers may pass pooled buffers; keep a copy.
	c.ectx.Request.Path = bytes.Clone(path)
}

func (c *Context) SetHost(host []byte) {
	sHost := bytesconv.B2s(host)
	c.req.Host = sHost

	if c.req.URL != nil {
		c.req.URL.Host = sHost
	}

	c.ectx.Request.Host = sHost
}

// SetQuery implements szhttp.Context.
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

// SetResponseHeader implements szhttp.Context.
func (c *Context) SetResponseHeader(key []byte, value []byte) {
	c.resp.Header().Set(string(key), string(value))
}

// SetStatusCode implements szhttp.Context.
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

// StatusCode implements szhttp.Context.
func (c *Context) StatusCode() int {
	return int(c.ectx.GetRequest().GetStatus())
}

// TLS implements szhttp.Context.
func (c *Context) TLS() bool {
	return c.Scheme() == httpconst.SchemeSecure
}

// URI implements szhttp.Context.
func (c *Context) URI() []byte {
	return []byte(c.req.URL.String())
}

// Unwrap implements szhttp.Context.
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

func (c *Context) JSON(statusCode int, body []byte) error {
	c.SetResponseHeader(szbytestr.HeaderContentType, szbytestr.ValueAppJSON)
	c.SetStatusCode(statusCode)

	// Use a JSON encoder to write the data
	encoder := json.NewEncoder(c.resp)
	return encoder.Encode(body)
}

func (c *Context) String(statusCode int, msg []byte) error {
	c.SetResponseHeader(szbytestr.HeaderContentType, szbytestr.ValueTextPlain)
	c.SetStatusCode(statusCode)

	_, err := c.resp.Write(msg)

	return err
}

func (c *Context) Render(statusCode int, component templ.Component) error {
	var buf bytes.Buffer
	if err := component.Render(c.Context(), &buf); err != nil {
		return err
	}

	c.SetResponseHeader(szbytestr.HeaderContentType, szbytestr.ValueTextHTML)
	c.SetStatusCode(statusCode)

	_, err := c.resp.Write(buf.Bytes())
	return err
	//return component.Render(c.Context(), c.resp)
}

func (c *Context) Upgrade() (*websocket.Conn, error) {
	conn, err := _upgrade.Upgrade(c.resp, c.req, nil)
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

// captureBody keeps a copy of a proxied response body for ResponseBody.
func (c *Context) captureBody(b []byte) {
	if c.bodyOverflow {
		return
	}

	if c.respBodyBuf.Len()+len(b) > _maxCapturedBody {
		c.bodyOverflow = true
		c.respBodyBuf.Reset()

		return
	}

	_, _ = c.respBodyBuf.Write(b)
}

func (c *Context) Request() *http.Request {
	return c.req
}

func (c *Context) Response() http.ResponseWriter {
	return c.resp
}

func (c *Context) Close() error {
	c.Reset()

	c.req = nil
	c.resp = nil

	c.ectx.Transport = nil
	c.ectx.Metadata = nil

	c.ectx.X.Reset()
	c.ectx.Request.Reset()
	c.respBodyBuf.Reset()

	c.bodyOverflow = false
	if c.respBodyBuf.Cap() > _maxPooledBodyBuf {
		c.respBodyBuf = bytes.Buffer{}
	}

	_ctxPool.Put(c)

	return nil
}
