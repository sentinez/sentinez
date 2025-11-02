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

package networks

import (
	"context"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
)

type RequestContext struct {
	Req *edgepb.RequestContext
	Ctx context.Context
}

func NewContext(ctx context.Context, req *edgepb.RequestContext) Context {
	return &RequestContext{
		Req: req,
		Ctx: ctx,
	}
}

func (c *RequestContext) GetReqProtocol() string {
	return c.Req.GetProtocol()
}

func (c *RequestContext) RemoteAddress() string {
	return c.Req.GetRemoteAddress()
}

func (c *RequestContext) StatusCode() int {
	return int(c.Req.GetStatusCode())
}

func (c *RequestContext) URI() string {
	return c.Req.GetUri()
}

func (c *RequestContext) Header() map[string]string {
	return c.Req.GetHeader()
}

func (c *RequestContext) Host() string {
	return c.Req.GetHost()
}

func (c *RequestContext) JA4() string {
	return c.Req.GetJa4()
}

func (c *RequestContext) Method() string {
	return c.Req.GetMethod()
}

func (c *RequestContext) Path() string {
	return c.Req.GetPath()
}

func (c *RequestContext) Queries() []string {
	return c.Req.GetQueries()
}

func (c *RequestContext) TLS() bool {
	return c.Req.GetTls()
}

func (c *RequestContext) Body() []byte {
	return c.Req.GetBody()
}

func (c *RequestContext) Context() context.Context {
	return c.Ctx
}

func (c *RequestContext) GetHeader() map[string]string {
	return c.Req.GetHeader()
}

func (c *RequestContext) GetHost() string {
	return c.Req.GetHost()
}

func (c *RequestContext) ClientIP() string {
	return c.Req.GetIp()
}

func (c *RequestContext) GetMethod() string {
	return c.Req.GetMethod()
}

func (c *RequestContext) GetPath() string {
	return c.Req.GetPath()
}

func (c *RequestContext) GetQueries() []string {
	return c.Req.GetQueries()
}
