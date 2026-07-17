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
	"bytes"
	"context"
	"net/url"
	"strings"
	"time"

	corehttp "github.com/sentinez/core/http"
	httpconst "github.com/sentinez/core/http/const"
	httppb "github.com/sentinez/sentinez/api/gen/go/sentinez/network/http/v1"
	"github.com/sentinez/shared/bytesconv"
)

func NewRequestContext(
	ctx context.Context, req *httppb.Request) corehttp.RequestContext {
	return &RequestContext{
		req: req,
		ctx: ctx,
	}
}

type RequestContext struct {
	req *httppb.Request
	ctx context.Context
}

// SetRequestId implements corehttp.RequestContext.
func (c *RequestContext) SetRequestId(id string) {
	c.req.Id = id
}

func (c *RequestContext) RequestTime() time.Time {
	return c.req.GetTimestamp().AsTime()
}

func (c *RequestContext) Header(k []byte) []byte {
	for _, header := range c.req.GetHeaders() {
		if bytes.Equal(header.GetKey(), k) {
			if len(header.GetValues()) == 0 {
				return nil
			}

			return header.GetValues()[0]
		}
	}

	return nil
}

func (c *RequestContext) Query(k []byte) []byte {
	for _, query := range c.req.GetQueries() {
		if bytes.Equal(query.GetKey(), k) {
			if len(query.GetValues()) == 0 {
				return nil
			}

			return query.GetValues()[0]
		}
	}

	return nil
}

func (c *RequestContext) QueryStr() []byte {
	var builder strings.Builder
	for _, query := range c.req.GetQueries() {
		for _, value := range query.GetValues() {
			if builder.Len() > 0 {
				builder.WriteByte('&')
			}
			builder.Write(query.GetKey())
			builder.WriteByte('=')
			builder.Write(value)
		}
	}
	return bytesconv.S2b(builder.String())
}

func (c *RequestContext) RemoteAddr() []byte {
	return bytesconv.S2b(c.req.GetRemoteAddress())
}

func (c *RequestContext) RequestId() string {
	return c.req.GetId()
}

func (c *RequestContext) SetBody(b []byte) {
	c.req.Body = b
}

func (c *RequestContext) SetRequestIP(ip []byte) {
	c.req.ClientIp = bytesconv.B2s(ip)
}

func (c *RequestContext) SetHeader(key, value []byte) {
	var current *httppb.RequestHeader

	for _, header := range c.req.GetHeaders() {
		if bytes.Equal(header.Key, key) {
			current = header
			break
		}
	}

	if current == nil {
		current = &httppb.RequestHeader{
			Key: key,
		}
		c.req.Headers = append(c.req.Headers, current)
	}

	current.Values = current.Values[:0]
	current.Values = append(current.Values, value)
}

func (c *RequestContext) SetHost(h []byte) {
	c.req.Host = bytesconv.B2s(h)
}

func (c *RequestContext) SetJA4(fingerprint string) {
	c.req.Fingerprint = fingerprint
}

func (c *RequestContext) SetMethod(m []byte) {
	c.req.Method = bytesconv.B2s(m)
}

func (c *RequestContext) SetPath(p []byte) {
	c.req.Path = p
}

func (c *RequestContext) SetProtocol(p string) {
	c.req.Protocol = p
}

func (c *RequestContext) SetQuery(key []byte, values ...[]byte) {
	var current *httppb.RequestQuery

	for _, query := range c.req.GetQueries() {
		if bytes.Equal(query.GetKey(), key) {
			current = query
			break
		}
	}

	if current == nil {
		current = &httppb.RequestQuery{
			Key: key,
		}
		c.req.Queries = append(c.req.Queries, current)
	}

	current.Values = current.Values[:0]
	current.Values = append(current.Values, values...)
}

func (c *RequestContext) SetRemoteAddr(addr []byte) {
	c.req.RemoteAddress = bytesconv.B2s(addr)
}

func (c *RequestContext) SetStatusCode(code int) {
	c.req.Status = int32(code)
}

func (c *RequestContext) SetURI(u []byte) {
	pURL, err := url.Parse(bytesconv.B2s(u))
	if err != nil {
		return
	}

	c.req.Uri = bytesconv.S2b(pURL.String())
}

func (c *RequestContext) Protocol() string {
	return c.req.GetProtocol()
}

func (c *RequestContext) StatusCode() int {
	return int(c.req.GetStatus())
}

func (c *RequestContext) URI() []byte {
	return c.req.GetUri()
}

func (c *RequestContext) Headers() map[string][][]byte {
	headers := make(map[string][][]byte)
	for _, header := range c.req.GetHeaders() {
		for _, value := range header.GetValues() {
			headers[bytesconv.B2s(header.GetKey())] =
				append(headers[bytesconv.B2s(header.GetKey())], value)
		}
	}

	return headers
}

func (c *RequestContext) Host() []byte {
	return bytesconv.S2b(c.req.GetHost())
}

func (c *RequestContext) JA4() string {
	return c.req.GetFingerprint()
}

func (c *RequestContext) Method() []byte {
	return bytesconv.S2b(c.req.GetMethod())
}

func (c *RequestContext) Path() []byte {
	return c.req.GetPath()
}

func (c *RequestContext) Queries() map[string][][]byte {
	queries := make(map[string][][]byte)
	for _, query := range c.req.GetQueries() {
		for _, value := range query.GetValues() {
			queries[bytesconv.B2s(query.GetKey())] =
				append(queries[bytesconv.B2s(query.GetKey())], value)
		}
	}

	return queries
}

func (c *RequestContext) TLS() bool {
	return c.Scheme() == httpconst.SchemeSecure
}

func (c *RequestContext) Body() []byte {
	return c.req.GetBody()
}

func (c *RequestContext) Context() context.Context {
	return c.ctx
}

func (c *RequestContext) RequestIP() []byte {
	return bytesconv.S2b(c.req.GetClientIp())
}

func (c *RequestContext) Scheme() string {
	return c.req.GetScheme()
}
