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

package enginectx

import (
	"context"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/core/networks"
	"github.com/sentinez/sentinez/pkg/x/syncx"
)

var (
	_ networks.Context = (*Context)(nil)

	pool = syncx.NewPool[Context]()
)

func New(ctx context.Context, ectx *edgepb.EngineContext) *Context {
	enginectx := pool.Get()
	enginectx.ectx = ectx
	enginectx.ctx = ctx
	return enginectx
}

type Context struct {
	ectx *edgepb.EngineContext
	ctx  context.Context
}

// GetReqProtocol implements networks.Context.
func (c *Context) GetReqProtocol() string {
	panic("unimplemented")
}

// RemoteAddress implements networks.Context.
func (c *Context) RemoteAddress() string {
	panic("unimplemented")
}

// StatusCode implements networks.Context.
func (c *Context) StatusCode() int {
	panic("unimplemented")
}

// URI implements networks.Context.
func (c *Context) URI() string {
	panic("unimplemented")
}

// Header implements networks.Context.
func (c *Context) Header() map[string]string {
	return c.ectx.GetHeader()
}

// Host implements networks.Context.
func (c *Context) Host() string {
	return c.ectx.GetHost()
}

// JA4 implements networks.Context.
func (c *Context) JA4() string {
	return c.ectx.GetJa4()
}

// Method implements networks.Context.
func (c *Context) Method() string {
	return c.ectx.GetMethod()
}

// Path implements networks.Context.
func (c *Context) Path() string {
	return c.ectx.GetPath()
}

// Queries implements networks.Context.
func (c *Context) Queries() []string {
	return c.ectx.GetQueries()
}

// TLS implements networks.Context.
func (c *Context) TLS() bool {
	return c.ectx.GetTls()
}

func (c *Context) Release() {
	c.ectx = nil
	c.ctx = nil
	pool.Put(c)
}

// Body implements context.Context.
func (c *Context) Body() []byte {
	return c.ectx.GetBody()
}

// Context implements context.Context.
func (c *Context) Context() context.Context {
	return c.ctx
}

// GetHeader implements context.Context.
func (c *Context) GetHeader() map[string]string {
	return c.ectx.GetHeader()
}

// GetHost implements context.Context.
func (c *Context) GetHost() string {
	return c.ectx.GetHost()
}

// ClientIP implements context.Context.
func (c *Context) ClientIP() string {
	return c.ectx.GetIp()
}

// GetMethod implements context.Context.
func (c *Context) GetMethod() string {
	return c.ectx.GetMethod()
}

// GetPath implements context.Context.
func (c *Context) GetPath() string {
	return c.ectx.GetPath()
}

// GetQueries implements context.Context.
func (c *Context) GetQueries() []string {
	return c.ectx.GetQueries()
}

// GetTLS implements context.Context.
func (c *Context) GetTLS() bool {
	return c.ectx.GetTls()
}
