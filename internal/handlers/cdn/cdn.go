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
	"time"

	"github.com/sentinez/core/szhttp/chains"

	"github.com/sentinez/contrib/pkg/bytesconv"
	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szstorage/cache/mem"
	"github.com/sentinez/pkg/szsync"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/memory/cdnrules"
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

// Handle serves and stores the responses any active CDN rule of the
// namespace matches, tried in priority order. System paths are answered
// by the edge itself and never cached.
func (c *Cache) Handle(ctx szhttp.Context, next chains.Handler) error {
	if bytes.HasPrefix(ctx.Path(), szbytestr.PathSystemPrefix) {
		return next.Handle(ctx)
	}

	if !matchAny(ctx, c.store.LoadContext(ctx)) {
		return next.Handle(ctx)
	}

	keyBuf := c.makeKey(ctx)
	defer func() {
		keyBuf.Reset()
		c.buf.Put(keyBuf)
	}()

	return c.caching(ctx, next, keyBuf.Bytes())
}

// matchAny reports whether a rule of chain matches ctx.
func matchAny(ctx szhttp.Context, chain []cdnrules.Entry) bool {
	for _, e := range chain {
		if e.Eval(ctx, nil) {
			return true
		}
	}

	return false
}

// caching serves the response cached under key, else stores the one of
// next. key is a pooled buffer: it is only copied when stored.
func (c *Cache) caching(
	ctx szhttp.Context, next chains.Handler, key []byte,
) error {

	resp, ok := c.cached.Get(bytesconv.B2s(key))
	if !ok {
		if err := next.Handle(ctx); err != nil {
			return err
		}

		c.setCache(ctx, key)
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
	// SetBody copies the body: the cached one is shared, never written.
	ctx.SetBody(resp.Body)

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

func (c *Cache) setCache(ctx szhttp.Context, key []byte) {
	// checked before the headers are copied out of the response
	if ctx.StatusCode() != http.StatusOK || len(ctx.ResponseBody()) == 0 {
		return
	}

	headers := ctx.ResponseHeader()
	if !cacheable(headers) {
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

	c.cached.Set(string(key), cr)
}

// cacheable reports whether a response, with headers, can be replayed to
// other clients: no cookie, no private or uncacheable directive. The
// status and the body are checked by setCache.
func cacheable(headers map[string][][]byte) bool {
	if len(headers[httpconst.HeaderSetCookie]) > 0 {
		return false
	}

	for _, v := range headers[httpconst.HeaderCacheControl] {
		if containsFold(v, "no-store") ||
			containsFold(v, "private") ||
			containsFold(v, "no-cache") {
			return false
		}
	}

	return true
}

// containsFold reports whether s, lower case ASCII, is in v whatever its
// case: a Cache-Control directive, matched without copying v.
func containsFold(v []byte, s string) bool {
	sb := bytesconv.S2b(s)
	for i := 0; i+len(sb) <= len(v); i++ {
		if bytes.EqualFold(v[i:i+len(sb)], sb) {
			return true
		}
	}

	return false
}
