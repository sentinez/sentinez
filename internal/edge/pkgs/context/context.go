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
	"github.com/sentinez/sentinez/core"
	"github.com/sentinez/sentinez/pkg/x/syncx"
)

var (
	_ core.RequestContext = (*Context)(nil)

	pool = syncx.NewPool[Context]()
)

func New(ctx context.Context, ectx *edgepb.EngineContext) *Context {
	enginectx := pool.Get()
	enginectx.EngineContext = ectx
	enginectx.ctx = ctx
	return enginectx
}

type Context struct {
	*edgepb.EngineContext
	ctx context.Context
}

func (c *Context) Release() {
	c.EngineContext = nil
	c.ctx = nil
	pool.Put(c)
}

// GetBody implements context.Context.
func (c *Context) GetBody() []byte {
	return c.EngineContext.GetBody()
}

// GetContext implements context.Context.
func (c *Context) GetContext() context.Context {
	return c.ctx
}

// GetHeader implements context.Context.
func (c *Context) GetHeader() map[string]string {
	return c.EngineContext.GetHeader()
}

// GetHost implements context.Context.
func (c *Context) GetHost() string {
	return c.EngineContext.GetHost()
}

// GetIP implements context.Context.
func (c *Context) GetIP() string {
	return c.GetIp()
}

// GetJA4 implements context.Context.
func (c *Context) GetJA4() string {
	return c.GetJa4()
}

// GetMethod implements context.Context.
func (c *Context) GetMethod() string {
	return c.EngineContext.GetMethod()
}

// GetPath implements context.Context.
func (c *Context) GetPath() string {
	return c.EngineContext.GetPath()
}

// GetQueries implements context.Context.
func (c *Context) GetQueries() []string {
	return c.EngineContext.GetQueries()
}

// GetTLS implements context.Context.
func (c *Context) GetTLS() bool {
	return c.GetTls()
}
