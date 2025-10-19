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

package httpxhz

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sentinez/sentinez"
	rulectx "github.com/sentinez/sentinez/corerule/context"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/sentinez/sentinez/pkg/x/syncx"
	"github.com/sentinez/sentinez/pkg/x/uuidx"
)

var (
	_        httpx.Context   = (*Context)(nil)
	_        rulectx.Context = (*Context)(nil)
	oncePool sync.Once
	ctxPool  *syncx.Pool[Context]
)

func NewContext(ctx context.Context, c *app.RequestContext) *Context {
	oncePool.Do(func() {
		ctxPool = syncx.NewPool[Context]()
	})

	httpCtx := ctxPool.Get()

	httpCtx.RequestContext = c
	httpCtx.ctx = ctx

	return httpCtx
}

type RequestHandler func(ctx *Context) error

type Context struct {
	*app.RequestContext
	ctx context.Context
}

// GetBody implements rulectx.Context.
func (c *Context) GetBody() []byte {
	return c.Request.Body()
}

// GetContext implements rulectx.Context.
func (c *Context) GetContext() context.Context {
	return c.Context()
}

// GetHeader implements rulectx.Context.
// Subtle: this method shadows the
// method (*RequestContext).GetHeader of Context.RequestContext.
func (c *Context) GetHeader() map[string]string {
	headers := make(map[string]string)
	c.VisitAllHeaders(func(key, value []byte) {
		headers[(string(key))] = string(value)
	})

	return headers
}

// GetHost implements rulectx.Context.
func (c *Context) GetHost() string {
	return string(c.Request.Host())
}

// GetIP implements rulectx.Context.
func (c *Context) GetIP() string {
	return c.ClientIP()
}

// GetJA4 implements rulectx.Context.
func (c *Context) GetJA4() string {
	return ""
}

// GetMethod implements rulectx.Context.
func (c *Context) GetMethod() string {
	return string(c.Request.Method())
}

// GetPath implements rulectx.Context.
func (c *Context) GetPath() string {
	return string(c.Request.Path())
}

// GetQueries implements rulectx.Context.
func (c *Context) GetQueries() []string {
	var queries []string
	c.VisitAllQueryArgs(func(key, _ []byte) {
		queries = append(queries, string(key))
	})

	return queries
}

// GetTLS implements rulectx.Context.
func (c *Context) GetTLS() bool {
	return true
}

func (c *Context) Time() time.Time {
	t, ok := c.ctx.Value(sntzRequestHTTPTimeKey).(time.Time)
	if ok {
		return t
	}
	return time.Time{}
}

// Context implements HTTPContext.
func (c *Context) Context() context.Context {
	c.GetConn()
	return c.ctx
}

// JSON implements HTTPContext.
func (c *Context) JSON(statusCode int, body []byte) error {
	c.SetContentType("application/json")
	c.SetStatusCode(statusCode)

	_, err := c.Write(body)
	return err
}

// Path implements HTTPContext.
func (c *Context) Path() string {
	return string(c.RequestContext.Request.URI().PathOriginal())
}

// Release implements HTTPContext.
func (c *Context) Release() {
	c.RequestContext = nil
	c.ctx = nil
	ctxPool.Put(c)
}

// String implements HTTPContext.
func (c *Context) String(statusCode int, body string) error {
	c.SetContentType("text/plain; charset=utf-8")
	c.SetStatusCode(statusCode)

	_, err := c.WriteString(body)
	return err
}

func (c *Context) SetServer() {
	c.Response.Header.Set("Server", sentinez.Name)
}

func (c *Context) GetReqID() string {
	res, ok := c.ctx.Value(sntzRequestHTTPIDKey).(string)
	if !ok {
		return ""
	}

	return res
}

func (c *Context) Render(statusCode int, component templ.Component) error {
	c.SetStatusCode(statusCode)
	c.SetContentType("text/html; charset=utf-8")
	c.SetServer()

	return component.Render(c.Context(), c.Response.BodyWriter())
}

func setIdentifier(ctx context.Context) context.Context {
	// set request time
	ctx = context.WithValue(ctx, sntzRequestHTTPTimeKey, time.Now().UTC())

	// set request id
	id := uuidx.NewIDHex(sentinez.PrefixRequestID)
	return context.WithValue(ctx, sntzRequestHTTPIDKey, id)
}

// GenContextKey nolint:funlen
func GenContextKey(ctx *Context) string {
	method := string(ctx.Method())
	host := string(ctx.Host())
	path := string(ctx.Path())

	args := ctx.QueryArgs()
	var keys []string
	args.VisitAll(func(key, _ []byte) {
		keys = append(keys, string(key))
	})
	sort.Strings(keys)

	sortedQuery := ""
	for _, k := range keys {
		sortedQuery += fmt.Sprintf("%s=%s&", k, args.Peek(k))
	}

	ct := string(ctx.Request.Header.ContentType())

	body := ctx.Request.Body()
	if len(body) > 1024 {
		body = body[:1024]
	}
	bodyHash := ""
	if len(body) > 0 {
		sum := sha256.Sum256(body)
		bodyHash = hex.EncodeToString(sum[:])
	}

	rawKey := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		method, host, path, sortedQuery, ct, bodyHash,
	)

	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}
