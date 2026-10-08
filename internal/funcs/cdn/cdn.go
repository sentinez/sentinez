// Copyright 2026 Sentinéz Labs.
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

package cdn

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/sentinez/core/common/bytestr"
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	corerule "github.com/sentinez/core/rules"
	"github.com/sentinez/core/storage/cache/mem"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

const (
	_headerAcceptEncoding = "Accept-Encoding"
	_headerAltSvc         = "Alt-Svc"
	_headerCacheControl   = "Cache-Control"
	_headerSetCookie      = "Set-Cookie"
)

func NewCache(_ zlog.Level, store *memory.MemStore) corechains.ChainNode {
	return &Cache{
		store:  store,
		Node:   corechains.NewNode(),
		cached: mem.New[*CachedResponse](time.Hour, time.Hour),
		buf:    sync.NewPool[bytes.Buffer](),
	}
}

type Cache struct {
	*corechains.Node
	store *memory.MemStore

	cached *mem.Cache[*CachedResponse]
	buf    *sync.Pool[bytes.Buffer]
}

func (c *Cache) Handle(ctx corehttp.Context) error {
	eval, rule := c.store.CDNRules().LoadContext(ctx)
	if rule.GetRuleRuntime().GetStatus() != typepb.Status_STATUS_ACTIVE {
		return c.HandleNext(ctx)
	}

	if !eval(ctx, nil) {
		return c.HandleNext(ctx)
	}

	keyBuf := c.makeKey(ctx)
	defer func() {
		keyBuf.Reset()
		c.buf.Put(keyBuf)
	}()

	return c.caching(ctx, eval, keyBuf.String())
}

func (c *Cache) caching(
	ctx corehttp.Context, eval corerule.EvalFunc, key string) error {

	resp, ok := c.cached.Get(key)
	if !ok {
		if err := c.HandleNext(ctx); err != nil {
			return err
		}

		c.setCache(ctx, eval, key)
		return nil
	}

	for k, vs := range resp.Headers {
		for _, v := range vs {
			ctx.AddResponseHeader(bytesconv.S2b(k), v)
		}
	}

	ctx.SetResponseHeader(bytestr.HeaderXCache, bytestr.HitCache)
	ctx.SetResponseHeader(bytestr.HeaderServer, bytestr.DefaultServerName)
	// Status must be written before the body, or it is silently dropped.
	ctx.SetStatusCode(resp.StatusCode)
	ctx.SetBody(bytes.Clone(resp.Body))

	return nil
}

type CachedResponse struct {
	StatusCode int
	Headers    map[string][][]byte
	Body       []byte
}

func (c *Cache) makeKey(ctx corehttp.Context) *bytes.Buffer {
	keyBuffer := c.buf.Get()
	_, _ = keyBuffer.Write(ctx.Method())
	_, _ = keyBuffer.WriteString(ctx.Scheme())
	_, _ = keyBuffer.Write(ctx.Host())
	_, _ = keyBuffer.Write(ctx.Path())
	_, _ = keyBuffer.Write(ctx.QueryStr())
	// The cached body is stored as encoded by the upstream (gzip, br, ...).
	_, _ = keyBuffer.Write(ctx.Header([]byte(_headerAcceptEncoding)))

	return keyBuffer
}

func (c *Cache) setCache(
	ctx corehttp.Context, ef corerule.EvalFunc, key string) {
	if ef == nil {
		return
	}

	headers := ctx.ResponseHeader()
	if !cacheable(ctx, headers) {
		return
	}

	cr := &CachedResponse{
		StatusCode: ctx.StatusCode(),
		Body:       bytes.Clone(ctx.ResponseBody()),
		Headers:    make(map[string][][]byte),
	}

	for k, vs := range headers {
		// Alt-Svc describes the edge listener, not the content; the server
		// sets it on every response, so caching it would duplicate it.
		if k == _headerAltSvc {
			continue
		}

		for _, v := range vs {
			cr.Headers[k] = append(cr.Headers[k], bytes.Clone(v))
		}
	}

	c.cached.Set(key, cr)
}

// cacheable reports whether a response can be replayed to other clients.
// Only complete 200 responses are stored: 206/304 bodies are partial or
// empty, and an empty body means it was not captured (e.g. too large).
func cacheable(ctx corehttp.Context, headers map[string][][]byte) bool {
	if ctx.StatusCode() != http.StatusOK || len(ctx.ResponseBody()) == 0 {
		return false
	}

	if len(headers[_headerSetCookie]) > 0 {
		return false
	}

	for _, v := range headers[_headerCacheControl] {
		cc := strings.ToLower(string(v))
		if strings.Contains(cc, "no-store") ||
			strings.Contains(cc, "private") ||
			strings.Contains(cc, "no-cache") {
			return false
		}
	}

	return true
}
