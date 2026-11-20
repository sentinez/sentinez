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

package script

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/scripting/szlua"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/internal/memory/scripts"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/stretchr/testify/require"
)

// BenchmarkHooksOneStage a namespace whose script hooks one stage, the
// request going through every stage scripts may hook.
func BenchmarkHooksOneStage(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	store := scripts.New()
	require.NoError(b, store.Store(_ns, []*edgepb.Script{{
		Id: "s1",
		Source: `
---@stage http:request/sec_rule
function noop(ctx) end`,
	}}))

	hooks := New(store)
	var h chains.Handler = chains.HandlerFunc(
		func(szhttp.Context) error { return nil })

	stages := szlua.Stages()
	for i := len(stages) - 1; i >= 0; i-- {
		h = hooks.Wrap(stages[i], h)
	}

	req := httptest.NewRequest(http.MethodGet, "http://"+_ns+"/api", nil)
	ctx := httpx.NewContext(req, httptest.NewRecorder())
	ctx.X().Namespace = _ns
	b.ReportAllocs()

	for b.Loop() {
		_ = h.Handle(ctx)
	}
}
