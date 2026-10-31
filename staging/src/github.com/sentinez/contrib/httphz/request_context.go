// Copyright 2026 Duc-Hung Ho.
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

package httphz

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sentinez/core/szcontext"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/szprefix"
	"github.com/sentinez/shared/szrand"
	"github.com/sentinez/shared/szstore/ja4"
	"github.com/sentinez/shared/szsync"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ szhttp.Context = (*Context)(nil)
	_ io.Closer      = (*Context)(nil)

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

func NewContext(ctx context.Context, c *app.RequestContext) *Context {
	hctx := _ctxPool.Get()
	hctx.SetRequestId(szrand.NewULID(szprefix.Request))
	hctx.req = c
	hctx.ctx = ctx

	transport := szcontext.GetTransport(ctx)
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
	req *app.RequestContext
	ctx context.Context

	ectx *edgepb.Context
}

// X implements [szhttp.Context].
func (c *Context) X() *edgepb.ContextExtra {
	return c.ectx.GetX()
}

func (c *Context) SetHost(h []byte) {
	c.req.Request.SetHost(bytesToString(h))
}

func (c *Context) AddResponseHeader(key, value []byte) {
	c.req.Response.Header.Add(bytesToString(key), bytesToString(value))
}

func (c *Context) Flush() error {
	return c.req.Flush()
}

func (c *Context) Header(k []byte) []byte {
	return c.req.GetHeader(bytesToString(k))
}

func (c *Context) Query(k []byte) []byte {
	return []byte(c.req.Query(bytesToString(k)))
}

func (c *Context) RemoteAddr() []byte {
	return []byte(c.req.RemoteAddr().String())
}

func (c *Context) ResponseBody() []byte {
	return c.req.Response.Body()
}

func (c *Context) ResponseHeader() map[string][][]byte {
	headers := make(map[string][][]byte)
	c.req.Response.Header.VisitAll(func(key, value []byte) {
		headers[bytesToString(key)] = append(headers[bytesToString(key)], value)
	})

	return headers
}

func (c *Context) Scheme() string {
	return bytesToString(c.req.Request.Scheme())
}

func (c *Context) SetHeader(key, value []byte) {
	c.req.Request.Header.Set(bytesToString(key), bytesToString(value))
}

func (c *Context) SetQuery(key []byte, values ...[]byte) {
	for _, v := range values {
		c.req.Request.URI().QueryArgs().
			Add(bytesToString(key), bytesToString(v))
	}
}

func (c *Context) SetResponseHeader(key, value []byte) {
	c.req.Response.Header.Set(bytesToString(key), bytesToString(value))
}

func (c *Context) SetURI(u []byte) {
	c.req.Request.SetRequestURI(bytesToString(u))
}

func (c *Context) QueryStr() []byte {
	return c.req.URI().QueryString()
}

func (c *Context) SetPath(path []byte) {
	c.req.URI().SetPath(bytesToString(path))
}

func (c *Context) Copy(src io.Reader) error {
	_, err := io.Copy(c.req, src)
	return err
}

func (c *Context) RequestBodyStream() io.Reader {
	return c.req.RequestBodyStream()
}

func (c *Context) VisitResponseHeaders(visitor func(k []byte, v []byte)) {
	c.req.Response.Header.VisitAll(func(key, value []byte) {
		visitor(key, value)
	})
}

func (c *Context) Protocol() string {
	return c.req.Request.Header.GetProtocol()
}

func (c *Context) ResetResponse() {
	c.req.Response.Reset()
}

func (c *Context) SetBody(body []byte) {
	c.req.Response.SetBody(body)
}

func (c *Context) SetStatusCode(code int) {
	c.req.SetStatusCode(code)
}

func (c *Context) StatusCode() int {
	return c.req.Response.StatusCode()
}

func (c *Context) URI() []byte {
	return []byte(c.req.URI().String())
}

func (c *Context) VisitRequestHeaders(visitor func(k []byte, v []byte)) {
	c.req.VisitAllHeaders(func(key, value []byte) {
		visitor(key, value)
	})
}

func (c *Context) Body() []byte {
	return c.req.Request.Body()
}

func (c *Context) GetContext() context.Context {
	return c.Context()
}

func (c *Context) Headers() map[string][][]byte {
	headers := make(map[string][][]byte)
	c.req.VisitAllHeaders(func(key, value []byte) {
		headers[bytesToString(key)] = append(headers[bytesToString(key)], value)
	})

	return headers
}

func (c *Context) Host() []byte {
	return c.req.Request.Host()
}

func (c *Context) RequestIP() []byte {
	return szhttp.StripPort([]byte(c.req.ClientIP()))
}

func (c *Context) JA4() string {
	return c.ectx.GetRequest().GetFingerprint()
}

func (c *Context) Method() []byte {
	return c.req.Request.Method()
}

func (c *Context) Queries() map[string][][]byte {
	params := make(map[string][][]byte)

	c.req.VisitAllQueryArgs(func(key, value []byte) {
		k := bytesToString(key)
		params[k] = append(params[k], value)
	})

	return params
}

func (c *Context) TLS() bool {
	return c.Scheme() == httpconst.SchemeSecure
}

func (c *Context) RequestTime() time.Time {
	return c.ectx.GetRequest().GetTimestamp().AsTime()
}

func (c *Context) Context() context.Context {
	c.req.GetConn()
	return c.ctx
}

func (c *Context) JSON(statusCode int, body []byte) error {
	c.req.SetContentType(httpconst.ValueAppJSON)
	c.req.SetStatusCode(statusCode)

	_, err := c.req.Write(body)
	return err
}

func (c *Context) File(path string) error {
	c.req.File(path)
	return nil
}

func (c *Context) Path() []byte {
	return c.req.Request.URI().PathOriginal()
}

func (c *Context) String(statusCode int, body []byte) error {
	c.req.SetContentType(httpconst.ValueTextPlain)
	c.req.SetStatusCode(statusCode)

	_, err := c.req.Write(body)
	return err
}

func (c *Context) Render(statusCode int, component templ.Component) error {
	c.req.SetStatusCode(statusCode)
	c.req.SetContentType(httpconst.ValueTextHTML)

	return component.Render(c.Context(), c.req.Response.BodyWriter())
}

func (c *Context) Unwrap() any {
	return c.req
}

func (c *Context) Close() error {
	c.Reset()

	c.req = nil
	c.ctx = nil
	c.ectx.Metadata = nil
	c.ectx.Transport = nil

	c.ectx.X.Reset()
	c.ectx.Request.Reset()

	_ctxPool.Put(c)

	return nil
}
