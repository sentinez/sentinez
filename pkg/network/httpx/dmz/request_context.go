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

package httpxdmz

import (
	"context"
	"io"
	"time"

	"github.com/a-h/templ"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sentinez/sentinez"
	"github.com/sentinez/sentinez/pkg/common/syncx"
	httpxbase "github.com/sentinez/sentinez/pkg/network/httpx/base"
)

var (
	_       httpxbase.Context = (*Context)(nil)
	ctxPool                   = syncx.NewPool[Context]()
)

const (
	HeaderXRequest = "X-Request-ID"
)

func NewContext(ctx context.Context, c *app.RequestContext) *Context {

	httpCtx := ctxPool.Get()

	httpCtx.req = c
	httpCtx.ctx = ctx

	return httpCtx
}

type RequestHandler func(ctx *Context) error

type Context struct {
	req *app.RequestContext
	ctx context.Context
}

// Copy implements networks.XContext.
func (c *Context) Copy(src io.Reader) error {
	_, err := io.Copy(c.req, src)
	return err
}

// RequestBodyStream implements networks.XContext.
func (c *Context) RequestBodyStream() io.Reader {
	return c.req.RequestBodyStream()
}

// GetRespHeader implements networks.XContext.
func (c *Context) GetRespHeader(k string) string {
	return c.req.Response.Header.Get(k)
}

// VisitRespHeaders implements networks.XContext.
func (c *Context) VisitRespHeaders(visitor func(k []byte, v []byte)) {
	c.req.Response.Header.VisitAll(func(key, value []byte) {
		visitor(key, value)
	})
}

// GetReqHeader implements networks.Context.
func (c *Context) GetReqHeader(k string) string {
	return string(c.req.GetHeader(k))
}

// GetProtocol implements networks.Context.
func (c *Context) Protocol() string {
	return c.req.Request.Header.GetProtocol()
}

// RemoteAddress implements networks.Context.
func (c *Context) RemoteAddress() string {
	return c.req.RemoteAddr().String()
}

// ResetResponse implements networks.Context.
func (c *Context) ResetResponse() {
	c.req.Response.Reset()
}

// SetBody implements networks.Context.
func (c *Context) SetBody(body []byte) {
	c.req.Response.SetBody(body)
}

// SetStatusCode implements networks.Context.
func (c *Context) SetStatusCode(code int) {
	c.req.SetStatusCode(code)
}

// StatusCode implements networks.Context.
func (c *Context) StatusCode() int {
	return c.req.Response.StatusCode()
}

// URI implements networks.Context.
func (c *Context) URI() string {
	return c.req.URI().String()
}

// VisitHeaders implements networks.Context.
func (c *Context) VisitReqHeaders(visitor func(k []byte, v []byte)) {
	c.req.VisitAllHeaders(func(key, value []byte) {
		visitor(key, value)
	})
}

// Body implements rulectx.Context.
func (c *Context) Body() []byte {
	return c.req.Request.Body()
}

// GetContext implements rulectx.Context.
func (c *Context) GetContext() context.Context {
	return c.Context()
}

// Header implements rulectx.Context.
func (c *Context) Header() map[string]string {
	headers := make(map[string]string)
	c.req.VisitAllHeaders(func(key, value []byte) {
		headers[(string(key))] = string(value)
	})

	return headers
}

// Host implements rulectx.Context.
func (c *Context) Host() string {
	return string(c.req.Request.Host())
}

// GetIP implements rulectx.Context.
func (c *Context) ClientIP() string {
	return c.req.ClientIP()
}

// JA4 implements rulectx.Context.
func (c *Context) JA4() string {
	return ""
}

// Method implements rulectx.Context.
func (c *Context) Method() string {
	return string(c.req.Request.Method())
}

// GetQueries implements rulectx.Context.
func (c *Context) Queries() map[string][]string {
	params := make(map[string][]string)

	c.req.VisitAllQueryArgs(func(key, value []byte) {
		k := string(key)
		v := string(value)
		params[k] = append(params[k], v)
	})

	return params
}

// TLS implements rulectx.Context.
func (c *Context) TLS() bool {
	return true
}

func (c *Context) Time() time.Time {
	t, ok := c.ctx.Value(senzRequestHTTPTimeKey).(time.Time)
	if ok {
		return t
	}
	return time.Time{}
}

// Context implements HTTPContext.
func (c *Context) Context() context.Context {
	c.req.GetConn()
	return c.ctx
}

// JSON implements HTTPContext.
func (c *Context) JSON(statusCode int, body []byte) error {
	c.req.SetContentType("application/json")
	c.req.SetStatusCode(statusCode)

	_, err := c.req.Write(body)
	return err
}

// Path implements HTTPContext.
func (c *Context) Path() string {
	return string(c.req.Request.URI().PathOriginal())
}

// Release implements HTTPContext.
func (c *Context) Release() {
	c.req = nil
	c.ctx = nil
	ctxPool.Put(c)
}

// String implements HTTPContext.
func (c *Context) String(statusCode int, body string) error {
	c.req.SetContentType("text/plain; charset=utf-8")
	c.req.SetStatusCode(statusCode)

	_, err := c.req.WriteString(body)
	return err
}

func (c *Context) SetServer() {
	c.req.Response.Header.Set("Server", sentinez.Name)
}

func (c *Context) GetReqID() string {
	return c.req.Request.Header.Get(HeaderXRequest)
}

func (c *Context) Render(statusCode int, component templ.Component) error {
	c.req.SetStatusCode(statusCode)
	c.req.SetContentType("text/html; charset=utf-8")
	c.SetServer()

	return component.Render(c.Context(), c.req.Response.BodyWriter())
}

func (c *Context) Unwrap() *app.RequestContext {
	return c.req
}
