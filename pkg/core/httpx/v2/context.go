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

// Package httpxv2 provides the HTTP context interface and its implementation.
package httpxv2

import (
	"github.com/sentinez/sentinez/pkg/core/httpx"
	"github.com/valyala/fasthttp"
)

// NewContext creates a new FastHTTP context.
// It implements the Context interface.
func NewContext(ctx *fasthttp.RequestCtx) Context {
	return &httpContext{
		core: ctx,
	}
}

type Context interface {
	httpx.Context
	AsCore() *fasthttp.RequestCtx
}

type httpContext struct {
	core *fasthttp.RequestCtx
}

func (c *httpContext) Method() string {
	return string(c.core.Method())
}

func (c *httpContext) AsCore() *fasthttp.RequestCtx {
	return c.core
}

func (c *httpContext) Path() string {
	return string(c.core.Request.URI().PathOriginal())
}

func (c *httpContext) String(statusCode int, body string) error {
	c.core.SetContentType("text/plain; charset=utf-8")
	c.core.SetStatusCode(statusCode)

	_, err := c.core.WriteString(body)

	return err
}

func (c *httpContext) JSON(statusCode int, body []byte) error {
	c.core.SetContentType("application/json")
	c.core.SetStatusCode(statusCode)

	_, err := c.core.Write(body)

	return err
}
