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

	"github.com/sentinez/core/szhttp/chains"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	"github.com/sentinez/core/szrule"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/memory/cdnrules"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/szstorage/cache/mem"
	"github.com/sentinez/shared/szsync"
)

// CacheId is the chain stage name of Cache.
const CacheId = "CDN"

var (
	_ chains.Plugin                      = (*Cache)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*Cache)(nil)
)

type Cache struct {
	store *cdnrules.Rule

	cached *mem.Cache[*CachedResponse]
	buf    *szsync.Pool[bytes.Buffer]
	log    szlog.LogCloser
}

func (*Cache) ID() string { return CacheId }

func (c *Cache) Provision(env *envplugin.Env) error {
	c.store = env.Store().CDNRules()
	c.cached = mem.New[*CachedResponse](time.Hour, time.Hour)
	c.buf = szsync.NewPool[bytes.Buffer]()
	c.log = env.Logger(typepb.LogType_LOG_TYPE_CDN_RULE)
	return nil
}

func (c *Cache) Handle(ctx szhttp.Context, next chains.Handler) error {
	eval, rule := c.store.LoadContext(ctx)
	if rule.GetRuleRuntime().GetStatus() != typepb.Status_STATUS_ACTIVE {
		return next.Handle(ctx)
	}

	if !eval(ctx, nil) {
		return next.Handle(ctx)
	}

	keyBuf := c.makeKey(ctx)
	defer func() {
		keyBuf.Reset()
		c.buf.Put(keyBuf)
	}()

	return c.caching(ctx, next, eval, keyBuf.String())
}

func (c *Cache) caching(ctx szhttp.Context, next chains.Handler,
	eval szrule.EvalFunc, key string) error {

	resp, ok := c.cached.Get(key)
	if !ok {
		if err := next.Handle(ctx); err != nil {
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

	ctx.SetResponseHeader(szbytestr.HeaderXCache, szbytestr.HitCache)
	ctx.SetResponseHeader(szbytestr.HeaderServer, szbytestr.DefaultServerName)
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

func (c *Cache) makeKey(ctx szhttp.Context) *bytes.Buffer {
	keyBuffer := c.buf.Get()
	_, _ = keyBuffer.Write(ctx.Method())
	_, _ = keyBuffer.WriteString(ctx.Scheme())
	_, _ = keyBuffer.Write(ctx.Host())
	_, _ = keyBuffer.Write(ctx.Path())
	_, _ = keyBuffer.Write(ctx.QueryStr())
	// The cached body is stored as encoded by the upstream (gzip, br, ...).
	_, _ = keyBuffer.Write(ctx.Header(szbytestr.HeaderAcceptEncoding))

	return keyBuffer
}

func (c *Cache) setCache(
	ctx szhttp.Context, ef szrule.EvalFunc, key string) {
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
		if k == httpconst.HeaderAltSvc {
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
func cacheable(ctx szhttp.Context, headers map[string][][]byte) bool {
	if ctx.StatusCode() != http.StatusOK || len(ctx.ResponseBody()) == 0 {
		return false
	}

	if len(headers[httpconst.HeaderSetCookie]) > 0 {
		return false
	}

	for _, v := range headers[httpconst.HeaderCacheControl] {
		cc := strings.ToLower(string(v))
		if strings.Contains(cc, "no-store") ||
			strings.Contains(cc, "private") ||
			strings.Contains(cc, "no-cache") {
			return false
		}
	}

	return true
}
