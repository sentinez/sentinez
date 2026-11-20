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

package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/pkg/szlog"
)

// _sink keeps the benchmarked results alive, so the compiler cannot drop
// or stack allocate them.
var _sink []byte

// newBenchRequest a request with the headers of a browser.
func newBenchRequest() *http.Request {
	req := httptest.NewRequest(http.MethodGet,
		"https://a.example.com/x/y.js?a=1&b=2", nil)
	for _, h := range [...]string{
		"User-Agent", "Accept", "Accept-Language", "Accept-Encoding",
		"Cookie", "Referer", "Sec-Fetch-Mode", "Sec-Fetch-Site",
		"Cache-Control", "X-Forwarded-For",
	} {
		req.Header.Set(h, "203.0.113.7")
	}

	return req
}

func BenchmarkContextLifecycle(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	req, rec := newBenchRequest(), httptest.NewRecorder()
	b.ReportAllocs()

	for b.Loop() {
		_ = NewContext(req, rec).Close()
	}
}

func BenchmarkContextGetters(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	ctx := NewContext(newBenchRequest(), httptest.NewRecorder())
	b.ReportAllocs()

	for b.Loop() {
		_sink = ctx.Host()
		_sink = ctx.Path()
		_sink = ctx.Method()
		_sink = ctx.QueryStr()
		_sink = ctx.URI()
		_sink = ctx.RemoteAddr()
		_sink = ctx.RequestIP()
		_sink = ctx.Header(szbytestr.HeaderUserAgent)
	}
}

func BenchmarkContextVisitRequestHeaders(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	ctx := NewContext(newBenchRequest(), httptest.NewRecorder())
	b.ReportAllocs()

	for b.Loop() {
		ctx.VisitRequestHeaders(func(_, v []byte) { _sink = v })
	}
}

func BenchmarkContextSetResponseHeader(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	ctx := NewContext(newBenchRequest(), httptest.NewRecorder())
	v := []byte("public, max-age=3600, immutable")
	b.ReportAllocs()

	for b.Loop() {
		ctx.SetResponseHeader(szbytestr.HeaderCacheControl, v)
	}
}
