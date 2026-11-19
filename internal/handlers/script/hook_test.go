// Copyright 2026 Duc-Hung Ho.
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

package script

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/internal/memory/scripts"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _ns = "script.example.com"

type fixture struct {
	rec    *httptest.ResponseRecorder
	ctx    *httpx.Context
	called bool
	next   chains.Handler
}

func newFixture(t *testing.T, ns string) *fixture {
	t.Helper()

	f := &fixture{rec: httptest.NewRecorder()}
	req := httptest.NewRequest(http.MethodGet, "http://"+_ns+"/api", nil)
	f.ctx = httpx.NewContext(req, f.rec)
	f.ctx.X().Namespace = ns
	f.next = chains.HandlerFunc(func(szhttp.Context) error {
		f.called = true

		return nil
	})

	return f
}

func storeOf(t *testing.T, source string) *scripts.Scripts {
	t.Helper()

	s := scripts.New()
	require.NoError(t, s.Store(_ns, []*edgepb.Script{
		{Id: "s1", Source: source},
	}))

	return s
}

func TestWrapSkipsStagesWithoutHooks(t *testing.T) {
	next := chains.HandlerFunc(func(szhttp.Context) error { return nil })
	h := New(scripts.New())

	for _, id := range []string{"TRC", "LOG", "DMA"} {
		wrapped := h.Wrap(id, next)
		assert.IsType(t, next, wrapped, id)
	}
}

func TestBeforeRespondEndsRequest(t *testing.T) {
	store := storeOf(t, `
---@module RUL
---@types before
function deny(ctx) ctx:respond(403, "no") end
`)
	f := newFixture(t, _ns)

	err := New(store).Wrap("RUL", f.next).Handle(f.ctx)

	require.NoError(t, err)
	assert.False(t, f.called, "the stage is skipped")
	assert.Equal(t, http.StatusForbidden, f.rec.Code)
	assert.Equal(t, "no", f.rec.Body.String())
}

func TestBeforeSetsHeaderAndContinues(t *testing.T) {
	store := storeOf(t, `
---@module ROU
---@types before
function tag(ctx) ctx:set_header("X-Edge", "sz") end
`)
	f := newFixture(t, _ns)

	err := New(store).Wrap("ROU", f.next).Handle(f.ctx)

	require.NoError(t, err)
	assert.True(t, f.called)
	assert.Equal(t, "sz", string(f.ctx.Header([]byte("X-Edge"))))
	assert.Equal(t, "sz", f.ctx.X().GetSetHeaders()["X-Edge"])
}

func TestHookOfAnotherStageDoesNotRun(t *testing.T) {
	store := storeOf(t, `
---@module RUL
---@types before
function deny(ctx) ctx:respond(403, "no") end
`)
	f := newFixture(t, _ns)

	err := New(store).Wrap("ROU", f.next).Handle(f.ctx)

	require.NoError(t, err)
	assert.True(t, f.called)
}

func TestOtherNamespaceHasNoHooks(t *testing.T) {
	store := storeOf(t, `
---@module RUL
---@types before
function deny(ctx) ctx:respond(403, "no") end
`)
	f := newFixture(t, "other.example.com")

	err := New(store).Wrap("RUL", f.next).Handle(f.ctx)

	require.NoError(t, err)
	assert.True(t, f.called)
}

func TestFailingScriptFailsOpen(t *testing.T) {
	store := storeOf(t, `
---@module RUL
---@types before
function boom(ctx) error("boom") end
`)
	f := newFixture(t, _ns)

	err := New(store).Wrap("RUL", f.next).Handle(f.ctx)

	require.NoError(t, err)
	assert.True(t, f.called)
}

func TestAfterRunsAfterStageAndCannotRespond(t *testing.T) {
	store := storeOf(t, `
---@module RUL
---@types after
function late(ctx) ctx:respond(500, "late") end
`)
	f := newFixture(t, _ns)

	err := New(store).Wrap("RUL", f.next).Handle(f.ctx)

	require.NoError(t, err)
	assert.True(t, f.called)
	assert.Empty(t, f.rec.Body.String())
}

func TestLogReachesTheEdge(t *testing.T) {
	store := storeOf(t, `
---@module ROU
---@types before
function f(ctx) log.info("hello", ctx:path()) log:debug("dbg") end
`)
	f := newFixture(t, _ns)

	err := New(store).Wrap("ROU", f.next).Handle(f.ctx)

	require.NoError(t, err)
	assert.True(t, f.called)
}
