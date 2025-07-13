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
	"github.com/sentinez/sentinez/pkg/core/httpx"
	"github.com/valyala/fasthttp"
)

// NewContext creates a new FastHTTP context.
// It implements the Context interface.
func NewContext(ctx *fasthttp.RequestCtx) *Context {
	return &Context{
		RequestCtx: ctx,
	}
}

type IContext interface {
	httpx.Context
}

type Context struct {
	*fasthttp.RequestCtx
}

func (c *Context) Path() string {
	return string(c.Request.URI().PathOriginal())
}

func (c *Context) String(statusCode int, body string) error {
	c.RequestCtx.SetContentType("text/plain; charset=utf-8")
	c.RequestCtx.SetStatusCode(statusCode)

	_, err := c.RequestCtx.WriteString(body)

	return err
}

func (c *Context) JSON(statusCode int, body []byte) error {
	c.RequestCtx.SetContentType("application/json")
	c.RequestCtx.SetStatusCode(statusCode)

	_, err := c.RequestCtx.Write(body)

	return err
}
