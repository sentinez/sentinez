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
	"testing"

	"github.com/sentinez/sentinez/internal/edge/v1/h/logging"
	"github.com/sentinez/sentinez/internal/edge/v1/h/routing"
	"github.com/sentinez/sentinez/internal/edge/v1/h/secure"
	"github.com/sentinez/sentinez/internal/edge/v1/h/static"
	"github.com/sentinez/sentinez/internal/edge/v1/h/waitingroom"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
	"github.com/sentinez/sentinez/shared/zlog"
)

func BenchmarkHandler(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet,
		"https://badcheese.is.s6z.io.vn:7443/", nil)
	w := httptest.NewRecorder()
	ctx := stdhttpx.NewContext(req, w)

	begin := waitingroom.New()

	zlog.SetLogLevel(zlog.LevelFatal.String())

	begin.SetNext(static.NewStatic()).
		SetNext(logging.NewLogger(zlog.LevelError)).
		SetNext(secure.NewDomain("is.s6z.io.vn")).
		SetNext(secure.NewWAF(zlog.LevelError)).
		SetNext(routing.NewMockRouter())

	b.ReportAllocs()

	for b.Loop() {
		_ = begin.Handle(ctx)
	}
}

func TestHandleChain(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"https://badcheese.is.s6z.io.vn:7443/", nil)
	w := httptest.NewRecorder()
	ctx := stdhttpx.NewContext(req, w)

	begin := waitingroom.New()
	begin.SetNext(static.NewStatic()).
		SetNext(logging.NewLogger(zlog.LevelInfo)).
		SetNext(secure.NewDomain("is.s6z.io.vn")).
		SetNext(secure.NewWAF(zlog.LevelInfo)).
		SetNext(routing.NewMockRouter())

	if err := begin.Handle(ctx); err != nil {
		t.Error(err)
	}
}
