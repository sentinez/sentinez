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

	"github.com/sentinez/core/szhttp/chains"

	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/handlers"
	"github.com/sentinez/sentinez/internal/handlers/logging"
	"github.com/sentinez/sentinez/internal/handlers/room"
	"github.com/sentinez/sentinez/internal/handlers/routing"
	"github.com/sentinez/sentinez/internal/handlers/secure"
	"github.com/sentinez/sentinez/internal/handlers/static"
	"github.com/sentinez/sentinez/internal/handlers/trace"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
	"github.com/sentinez/shared/szlog"
	"github.com/stretchr/testify/require"
)

var _reqPool = sync.Pool{
	New: func() any {
		return httptest.NewRequest(
			http.MethodGet, "https://badcheese.is.s6z.io.vn:7443/", nil)
	},
}

var _respPool = sync.Pool{
	New: func() any {
		return httptest.NewRecorder()
	},
}

func getRequest() *http.Request {
	req := _reqPool.Get().(*http.Request)
	return req
}

func putRequest(req *http.Request) {
	_reqPool.Put(req)
}

func getRecorder() *httptest.ResponseRecorder {
	return _respPool.Get().(*httptest.ResponseRecorder)
}

func putRecorder(w *httptest.ResponseRecorder) {
	_respPool.Put(w)
}

func BenchmarkStandardConverter(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	begin := newTestChain(b)

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
	ctx := httpx.NewContext(req, w)

	szlog.SetLogLevel(szlog.LevelInfo)

	begin := newTestChain(t)

	require.NoError(t, begin.Handle(ctx))
}

// newTestChain links the edge plugins without a memory store, ending
// with a mock router.
func newTestChain(tb testing.TB) *chains.Chain {
	tb.Helper()

	env := envplugin.NewEnv(nil,
		option.WithLevel(szlog.LevelError),
		option.WithHostname("is.s6z.io.vn"),
	)
	chain, err := chains.Build(env, handlers.Stage,
		&trace.Tracer{},
		&room.WaitingRoom{},
		&static.Static{},
		&logging.Logging{},
		&secure.DomainBased{},
		&secure.SecRule{},
		&secure.Rulesets{},
		&routing.Mock{},
	)
	require.NoError(tb, err)
	tb.Cleanup(func() { require.NoError(tb, chain.Close()) })

	return chain
}
