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

package corehttpreq

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	corehttp "github.com/sentinez/core/http"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
)

func NewRequestContext(
	ctx context.Context, req *edgepb.RequestContext) corehttp.RequestContext {
	return &RequestContext{
		Req: req,
		Ctx: ctx,
	}
}

type RequestContext struct {
	Req *edgepb.RequestContext
	Ctx context.Context
}

func (c *RequestContext) RequestTime() time.Time {
	return time.Now().UTC()
}

func (c *RequestContext) Header(k string) string {
	if h := c.Req.GetHeader(); h != nil {
		return h[k]
	}

	return ""
}

func (c *RequestContext) Query(k string) string {
	if q := c.Req.GetQueries(); q != nil {
		values := q[k]
		if values.GetValue() != nil {
			return values.GetValue()[0]
		}
	}

	return ""
}

func (c *RequestContext) QueryStr() string {
	result := ""
	for k, vs := range c.Req.GetQueries() {
		for _, value := range vs.GetValue() {
			result += fmt.Sprintf("%v=%v?", k, value)
		}
	}

	return strings.TrimSuffix(result, "?")
}

func (c *RequestContext) RemoteAddr() string {
	return c.Req.GetRemoteAddress()
}

func (c *RequestContext) RequestId() string {
	return c.Req.GetId()
}

func (c *RequestContext) SetBody(b []byte) {
	c.Req.Body = b
}

func (c *RequestContext) SetClientIP(ip string) {
	c.Req.Ip = ip
}

func (c *RequestContext) SetHeader(key string, value string) {
	if c.Req.Header == nil {
		c.Req.Header = make(map[string]string)
	}

	c.Req.Header[key] = value
}

func (c *RequestContext) SetHost(h string) {
	c.Req.Host = h
}

func (c *RequestContext) SetJA4(fingerprint string) {
	c.Req.Ja4 = fingerprint
}

func (c *RequestContext) SetMethod(m string) {
	c.Req.Method = m
}

func (c *RequestContext) SetPath(p string) {
	c.Req.Path = p
}

func (c *RequestContext) SetProtocol(p string) {
	c.Req.Protocol = p
}

func (c *RequestContext) SetQuery(key string, values ...string) {
	c.Req.Queries[key].Value = append(c.Req.Queries[key].Value, values...)
}

func (c *RequestContext) SetRemoteAddr(addr string) {
	c.Req.RemoteAddress = addr
}

func (c *RequestContext) SetStatusCode(code int) {
	c.Req.StatusCode = int32(code)
}

func (c *RequestContext) SetURI(u string) {
	c.Req.Uri = u
}

func (c *RequestContext) Protocol() string {
	return c.Req.GetProtocol()
}

func (c *RequestContext) StatusCode() int {
	return int(c.Req.GetStatusCode())
}

func (c *RequestContext) URI() string {
	return c.Req.GetUri()
}

func (c *RequestContext) Headers() map[string]string {
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

func (c *RequestContext) Queries() map[string][]string {
	params := make(map[string][]string)
	for k, v := range c.Req.GetQueries() {
		params[k] = v.GetValue()
	}

	return params
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

func (c *RequestContext) ClientIP() string {
	return c.Req.GetIp()
}

func (c *RequestContext) Scheme() string {
	u, err := url.Parse(c.Req.GetUri())
	if err != nil {
		panic(err)
	}

	return u.Scheme
}
