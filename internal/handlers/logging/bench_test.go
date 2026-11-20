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

package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/sentinez/sentinez/pkg/pools/egress"
	"github.com/sentinez/sentinez/pkg/pools/ingress"
)

// newBenchContext a proxied request with the headers of a browser,
// answered by the upstream.
func newBenchContext() szhttp.Context {
	req := httptest.NewRequest(http.MethodGet,
		"https://a.example.com/x/y.js?a=1&b=2", nil)
	for _, h := range [...]string{
		"User-Agent", "Accept", "Accept-Language", "Accept-Encoding",
		"Cookie", "Referer", "Sec-Fetch-Mode", "Sec-Fetch-Site",
		"Cache-Control", "Content-Type",
	} {
		req.Header.Set(h, "some value of the header")
	}

	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "text/javascript")
	rec.Header().Set("Content-Length", "1024")
	rec.Header().Set("Cache-Control", "public")

	ctx := httpx.NewContext(req, rec)
	ctx.X().Upstream = "http://10.0.0.1:8080"
	szhttp.SetUpstreamHeader(ctx, "X-Tenant", "t1")

	return ctx
}

func BenchmarkNewEvent(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	ctx := newBenchContext()
	b.ReportAllocs()

	for b.Loop() {
		ingress.Release(newEvent(ctx, ctx.RequestIP()))
	}
}

func BenchmarkNewEgressEvent(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	ctx := newBenchContext()
	b.ReportAllocs()

	for b.Loop() {
		egress.Release(newEgressEvent(ctx, ctx.X()))
	}
}
