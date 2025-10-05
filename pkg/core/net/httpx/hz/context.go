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

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sentinez/sentinez/pkg/core/net/httpx"
	"github.com/sentinez/sentinez/pkg/syncx"
	"github.com/sentinez/sentinez/pkg/uuid"
)

var (
	_        HTTPContext = (*Context)(nil)
	oncePool sync.Once
	ctxPool  *syncx.Pool[Context]
)

type HTTPContext interface {
	httpx.Context
	Context() context.Context
	Time() time.Time
}

func NewContext(ctx context.Context, c *app.RequestContext) *Context {
	oncePool.Do(func() {
		ctxPool = syncx.NewPool[Context]()
	})

	httpCtx := ctxPool.Get()

	httpCtx.RequestContext = c
	httpCtx.ctx = ctx

	return httpCtx
}

type Context struct {
	*app.RequestContext
	ctx context.Context
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

func setIdentifier(ctx context.Context) context.Context {
	id := uuid.NewHex("SNTZREQ")
	ctx = context.WithValue(ctx, sntzRequestHTTPTimeKey, time.Now())
	return context.WithValue(ctx, sntzRequestHTTPIDKey, id)
}

func GetContextIdentify(ctx *Context) string {
	res, ok := ctx.ctx.Value(sntzRequestHTTPIDKey).(string)
	if !ok {
		return ""
	}

	return res
}

// GenerateContextKey nolint:funlen
func GenerateContextKey(ctx *Context) string {
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
