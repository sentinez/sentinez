// Copyright 2026 Duc-Hung Ho.
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
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szstorage/cache/mem"
	"github.com/sentinez/pkg/szsync"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	cdnpb "github.com/sentinez/sentinez/api/proto/sentinez/types/cdn/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory/cdnrules"
	"github.com/sentinez/sentinez/pkg/network/httpx"
)

const _benchNamespace = "bench.example.com"

// discard a response writer dropping what is written: the benchmarks
// reuse it, so it must not grow.
type discard struct{ h http.Header }

func (d *discard) Header() http.Header       { return d.h }
func (*discard) Write(b []byte) (int, error) { return len(b), nil }
func (*discard) WriteHeader(int)             {}

// newBenchCache a Cache whose namespace has a rule matching every request.
func newBenchCache() *Cache {
	store := cdnrules.New()
	store.Store(_benchNamespace, &edgepb.Delivery{
		Cdn: []*cdnpb.CDN{{Rule: &cdnpb.Rule{
			Id:     "cdn_bench",
			Status: typepb.Status_STATUS_ACTIVE,
			Expr: &rulepb.Expression{
				OrCondition: []*rulepb.AndCondition{{}},
			},
		}}},
	})

	return &Cache{
		store:  store,
		cached: mem.New[*CachedResponse](time.Hour, time.Hour),
		buf:    szsync.NewPool[bytes.Buffer](),
	}
}

// newBenchContext a request of the namespace, answered into w.
func newBenchContext(w *discard) szhttp.Context {
	req := httptest.NewRequest(http.MethodGet,
		"https://"+_benchNamespace+"/app.js?v=1", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	ctx := httpx.NewContext(req, w)
	ctx.X().Namespace = _benchNamespace

	return ctx
}

// upstream answers with status, a body of 16 KiB and 6 headers.
func upstream(status int) chains.Handler {
	body := bytes.Repeat([]byte("x"), 16<<10)

	return chains.HandlerFunc(func(ctx szhttp.Context) error {
		for _, k := range [...]string{
			"Content-Type", "Etag", "Last-Modified", "Vary", "Date", "Age",
		} {
			ctx.SetResponseHeader([]byte(k), []byte("some value"))
		}

		ctx.SetStatusCode(status)
		ctx.SetBody(body)

		return nil
	})
}

func BenchmarkCacheHit(b *testing.B) {
	benchCache(b, http.StatusOK, true)
}

// BenchmarkCacheMissUncacheable a response the cache cannot keep, a 404:
// every request is a miss.
func BenchmarkCacheMissUncacheable(b *testing.B) {
	benchCache(b, http.StatusNotFound, false)
}

// benchCache serves b.N requests answered with status by the upstream,
// the first one filling the cache if warm.
func benchCache(b *testing.B, status int, warm bool) {
	szlog.SetLogLevel(szlog.LevelInfo)

	c, w := newBenchCache(), &discard{h: http.Header{}}
	ctx := newBenchContext(w)
	next := upstream(status)

	if warm {
		if err := c.Handle(ctx, next); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()

	for b.Loop() {
		clear(w.h)
		ctx.ResetResponse()
		_ = c.Handle(ctx, next)
	}
}
