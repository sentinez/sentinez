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

// Package httpxf1 provides the HTTP context interface and its implementation.
package httpxf1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"github.com/sentinez/sentinez"
	"github.com/sentinez/sentinez/pkg/common/syncx"
	"github.com/sentinez/sentinez/pkg/common/uuid"
	"github.com/sentinez/sentinez/pkg/core/net/httpx"
	"github.com/valyala/fasthttp"
)

const userValueKey = "sntz_request_hex"

var (
	oncePool sync.Once
	ctxPool  *syncx.Pool[Context]
)

// NewContext creates a new FastHTTP context.
// It implements the Context interface.
func NewContext(ctx *fasthttp.RequestCtx) *Context {
	oncePool.Do(func() {
		ctxPool = syncx.NewPool[Context]()
	})

	httpCtx := ctxPool.Get()

	httpCtx.RequestCtx = ctx

	return httpCtx
}

type HTTPContext interface {
	httpx.Context
}

type Context struct {
	*fasthttp.RequestCtx
}

func (c *Context) Release() {
	c.RequestCtx = nil
	ctxPool.Put(c)
}

func (c *Context) Path() string {
	return string(c.Request.URI().PathOriginal())
}

func (c *Context) String(statusCode int, body string) error {
	c.SetContentType("text/plain; charset=utf-8")
	c.SetStatusCode(statusCode)

	_, err := c.WriteString(body)

	return err
}

func (c *Context) JSON(statusCode int, body []byte) error {
	c.SetContentType("application/json")
	c.SetStatusCode(statusCode)

	_, err := c.Write(body)

	return err
}

func setIdentifier(ctx *fasthttp.RequestCtx) {
	id := uuid.NewIDHex(sentinez.PrefixRequestID)
	ctx.SetUserValue(userValueKey, id)
}

func GetContextIdentify(ctx *Context) string {
	res, ok := ctx.UserValue(userValueKey).(string)
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
	args.All()(func(k, _ []byte) bool {
		keys = append(keys, string(k))
		return true
	})
	sort.Strings(keys)

	sortedQuery := ""
	for _, k := range keys {
		sortedQuery += fmt.Sprintf("%s=%s&", k, args.Peek(k))
	}

	ct := string(ctx.Request.Header.ContentType())

	body := ctx.PostBody()
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
