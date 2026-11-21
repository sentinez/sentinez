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

package scripting

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sentinez/scripting/exchange"
	"github.com/sentinez/scripting/phase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeExchange struct {
	mu       sync.Mutex
	headers  map[string]string
	setHdr   map[string]string
	respHdr  map[string]string
	status   int
	body     string
	answered bool
	logs     []string
}

func newFake() *fakeExchange {
	return &fakeExchange{
		headers: map[string]string{"X-Debug": "1"},
		setHdr:  map[string]string{},
		respHdr: map[string]string{},
	}
}

func (f *fakeExchange) Method() []byte    { return []byte("GET") }
func (f *fakeExchange) Path() []byte      { return []byte("/api") }
func (f *fakeExchange) Host() []byte      { return []byte("example.com") }
func (f *fakeExchange) RequestIP() []byte { return []byte("10.0.0.1") }
func (f *fakeExchange) Query(k []byte) []byte {
	return []byte("q-" + string(k))
}

func (f *fakeExchange) Header(k []byte) []byte {
	return []byte(f.headers[string(k)])
}

func (f *fakeExchange) SetHeader(k, v string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.setHdr[k] = v
}

func (f *fakeExchange) SetResponseHeader(k, v string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.respHdr[k] = v
}

func (f *fakeExchange) Log(level exchange.Level, script, msg string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.logs = append(f.logs, level.String()+" "+script+": "+msg)
}

func (f *fakeExchange) String(status int, body []byte) error {
	f.status, f.body, f.answered = status, string(body), true

	return nil
}

func TestCompileInvalid(t *testing.T) {
	tests := []struct {
		name string
		give string
	}{
		{"syntax", "function ("},
		{"runtime error in body", "error('boom')"},
		{"os is not open", "os.exit(1)"},
		{"io is not open", "io.open('/etc/passwd')"},
		{"load is removed", "load('return 1')()"},
		{"require is removed", "require('os')"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compile("t.lua", tt.give)
			assert.Error(t, err)
		})
	}
}

func TestCallReadsAndSets(t *testing.T) {
	p, err := Compile("t.lua", `
---@stage http:request/router
function tag(ctx)
  if ctx:header("X-Debug") == "1" and ctx:path() == "/api" then
    ctx:set_header("X-Edge", ctx:method() .. ctx:query("a"))
    ctx:set_response_header("X-Seen", ctx:host() .. ctx:ip())
  end
end
`)
	require.NoError(t, err)

	ex := newFake()
	responded, err := p.Call(context.Background(), phase.Request, "ROU", ex)
	require.NoError(t, err)

	assert.False(t, responded)
	assert.Equal(t, "GETq-a", ex.setHdr["X-Edge"])
	assert.Equal(t, "example.com10.0.0.1", ex.respHdr["X-Seen"])
	assert.False(t, ex.answered)
}

func TestCallRespondStopsTheNextFunctions(t *testing.T) {
	p, err := Compile("t.lua", `
---@stage http:request/sec_rule
function deny(ctx) ctx:respond(403, "no") end

---@stage http:request/sec_rule
function never(ctx) ctx:set_header("X-Never", "1") end
`)
	require.NoError(t, err)

	ex := newFake()
	responded, err := p.Call(context.Background(), phase.Request, "RUL", ex)
	require.NoError(t, err)

	assert.True(t, responded)
	assert.Equal(t, 403, ex.status)
	assert.Equal(t, "no", ex.body)
	assert.Empty(t, ex.setHdr)
}

func TestCallRunsInSourceOrder(t *testing.T) {
	p, err := Compile("t.lua", `
---@stage http:response/sec_ruleset
function second(ctx) ctx:set_header("X-Order", ctx:header("X-Order") .. "2") end

---@stage http:response/sec_ruleset
function first(ctx) ctx:set_header("X-Order", "1") end
`)
	require.NoError(t, err)

	ex := newFake()
	ex.headers["X-Order"] = "0"
	_, err = p.Call(context.Background(), phase.Response, "RST", ex)
	require.NoError(t, err)

	// second runs first: it reads the request header, not what first set
	assert.Equal(t, "1", ex.setHdr["X-Order"])
}

func TestCallErrors(t *testing.T) {
	tests := []struct {
		name string
		give string
	}{
		{"lua error", "error('boom')"},
		{"bad status", "ctx:respond(42)"},
		{"bad argument", "ctx:header({})"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "---@stage http:request/static\n" +
				"function f(ctx) " + tt.give + " end"
			p, err := Compile("t.lua", src)
			require.NoError(t, err)

			ex := newFake()
			responded, err := p.Call(context.Background(),
				phase.Request, "STC", ex)
			assert.Error(t, err)
			assert.False(t, responded)
			assert.False(t, ex.answered)
		})
	}
}

func TestCallFunctionRedefinedToNil(t *testing.T) {
	p, err := Compile("t.lua", `
---@stage http:request/static
function f(ctx) f = nil end
`)
	require.NoError(t, err)

	_, err = p.Call(context.Background(), phase.Request, "STC", newFake())
	require.NoError(t, err)

	// the state that ran it was kept: f is gone from it
	_, err = p.Call(context.Background(), phase.Request, "STC", newFake())
	assert.ErrorContains(t, err, "f is not a function anymore")
}

func TestCallTimeout(t *testing.T) {
	p, err := Compile("t.lua", `
---@stage http:request/cache
function spin(ctx) while true do end end
`, WithTimeout(20*time.Millisecond))
	require.NoError(t, err)

	start := time.Now()
	_, err = p.Call(context.Background(), phase.Request, "CDN", newFake())

	require.Error(t, err)
	assert.Less(t, time.Since(start), 2*time.Second)

	// the program still works after a state was discarded
	_, err = p.Call(context.Background(), phase.Request, "CDN", newFake())
	assert.Error(t, err)
}

func TestCallWithoutHook(t *testing.T) {
	p, err := Compile("t.lua", `
---@stage http:request/cache
function f(ctx) ctx:respond(403) end
`)
	require.NoError(t, err)

	tests := []struct {
		name      string
		givePhase phase.Phase
		giveStage string
	}{
		{"other stage", phase.Request, "ROU"},
		{"other phase", phase.Response, "CDN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex := newFake()
			responded, err := p.Call(context.Background(), tt.givePhase,
				tt.giveStage, ex)
			require.NoError(t, err)
			assert.False(t, responded)
			assert.False(t, ex.answered)
		})
	}
}

func TestCallConcurrent(t *testing.T) {
	p, err := Compile("t.lua", `
---@stage http:request/sec_rate_limit
function f(ctx) ctx:set_header("X", ctx:path()) end
`)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range 32 {
		wg.Go(func() {
			for range 20 {
				ex := newFake()
				_, err := p.Call(context.Background(), phase.Request, "LMT", ex)
				assert.NoError(t, err)
				assert.Equal(t, "/api", ex.setHdr["X"])
			}
		})
	}

	wg.Wait()
}

func TestCallAfterCannotRespond(t *testing.T) {
	p, err := Compile("t.lua", `
---@stage http:response/sec_rule
function late(ctx) ctx:respond(500, "late") end
`)
	require.NoError(t, err)

	ex := newFake()
	responded, err := p.Call(context.Background(), phase.Response, "RUL", ex)

	assert.ErrorContains(t, err, "the stage already ran")
	assert.False(t, responded)
	assert.False(t, ex.answered)
}

func TestCallErrorNamesTheSection(t *testing.T) {
	p, err := Compile("t.lua", `
---@stage http:response/sec_ruleset
function boom(ctx) error("x") end
`)
	require.NoError(t, err)

	_, err = p.Call(context.Background(), phase.Response, "RST", newFake())
	assert.ErrorContains(t, err,
		"script t.lua: http:response/sec_ruleset: boom:")
}

func TestCallKeptCtxIsDead(t *testing.T) {
	p, err := Compile("t.lua", `
local kept
---@stage http:request/router
function keep(ctx) kept = ctx end
---@stage http:response/router
function reuse(ctx) kept:path() end
`)
	require.NoError(t, err)

	_, err = p.Call(context.Background(), phase.Request, "ROU", newFake())
	require.NoError(t, err)

	_, err = p.Call(context.Background(), phase.Response, "ROU", newFake())
	assert.ErrorContains(t, err, "ctx expected")
}
