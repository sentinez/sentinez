// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edge

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/sentinez/sentinez/internal/edge/v1/http/logging"
	"github.com/sentinez/sentinez/internal/edge/v1/http/room"
	"github.com/sentinez/sentinez/internal/edge/v1/http/routing"
	"github.com/sentinez/sentinez/internal/edge/v1/http/secure"
	"github.com/sentinez/sentinez/internal/edge/v1/http/static"
	"github.com/sentinez/sentinez/internal/edge/v1/http/trace"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
	"github.com/sentinez/shared/zlog"
)

var reqPool = sync.Pool{
	New: func() any {
		return httptest.NewRequest(
			http.MethodGet, "https://badcheese.is.s6z.io.vn:7443/", nil)
	},
}

var respPool = sync.Pool{
	New: func() any {
		return httptest.NewRecorder()
	},
}

func getRequest() *http.Request {
	req := reqPool.Get().(*http.Request)
	return req
}

func putRequest(req *http.Request) {
	reqPool.Put(req)
}

func getRecorder() *httptest.ResponseRecorder {
	return respPool.Get().(*httptest.ResponseRecorder)
}

func putRecorder(w *httptest.ResponseRecorder) {
	respPool.Put(w)
}

func BenchmarkStandardConverter(b *testing.B) {
	zlog.SetLogLevel(zlog.LevelInfo)

	begin := trace.NewTracer(zlog.LevelError)
	begin.
		SetNext(room.NewRoom(zlog.LevelError)).
		SetNext(static.NewStatic(zlog.LevelError)).
		SetNext(logging.NewLogger(zlog.LevelError)).
		SetNext(secure.NewDomainBased("is.s6z.io.vn")).
		SetNext(secure.NewRuleBased(zlog.LevelError)).
		SetNext(secure.NewWAF(zlog.LevelError)).
		SetNext(routing.NewMockRouter())

	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := getRequest()
			resp := getRecorder()

			stdhttpx.StandardConverter(begin.Handle, resp, req)

			putRequest(req)
			putRecorder(resp)
		}
	})
}

func TestHandleChain(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"https://badcheese.is.s6z.io.vn:7443/", nil)
	w := httptest.NewRecorder()
	ctx := stdhttpx.NewContext(req, w)

	zlog.SetLogLevel(zlog.LevelInfo)

	begin := trace.NewTracer(zlog.LevelError)
	begin.
		SetNext(room.NewRoom(zlog.LevelError)).
		SetNext(static.NewStatic(zlog.LevelError)).
		SetNext(logging.NewLogger(zlog.LevelError)).
		SetNext(secure.NewDomainBased("is.s6z.io.vn")).
		SetNext(secure.NewRuleBased(zlog.LevelError)).
		SetNext(secure.NewWAF(zlog.LevelError)).
		SetNext(routing.NewMockRouter())

	if err := begin.Handle(ctx); err != nil {
		t.Error(err)
	}
}
