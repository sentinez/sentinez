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

package szlua

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _hooksScript = `
-- any comment
---@stage http:request/sec_rule
---@param ctx sz.Context
---@stage http:request/sec_ruleset
function whatever_name(ctx) end

---@stage http:response/router
function on_router(ctx) end

function before_cache(ctx) end -- a name alone declares nothing

---@stage http:request/static
---@stage http:response/static
function both(ctx) end

---@module 'socket'
---@type table
local socket = {} -- other annotations: left alone
`

func TestCompileHooks(t *testing.T) {
	p, err := Compile("t.lua", _hooksScript)
	require.NoError(t, err)

	tests := []struct {
		givePhase Phase
		giveStage string
		want      bool
	}{
		{Request, "RUL", true},
		{Request, "RST", true},
		{Response, "ROU", true},
		{Request, "STC", true},
		{Response, "STC", true},
		{Request, "ROU", false},
		{Response, "RUL", false},
		{Request, "CDN", false},
	}

	for _, tt := range tests {
		name := tt.givePhase.String() + " " + tt.giveStage
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, p.Has(tt.givePhase, tt.giveStage))
		})
	}
}

type invalidAnnotationTest struct {
	name    string
	give    string
	wantErr string
}

func TestCompileInvalidAnnotation(t *testing.T) {
	for _, tt := range invalidAnnotationTests() {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compile("t.lua", tt.give)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func invalidAnnotationTests() []invalidAnnotationTest {
	return []invalidAnnotationTest{
		{"empty", "---@stage\nfunction f(c) end",
			`line 1: @stage takes phase/stage, ` +
				`e.g. http:request/sec_rule, got ""`},
		{"no slash", "---@stage http:request\nfunction f(c) end",
			"line 1: @stage takes phase/stage"},
		{"two stages", "---@stage http:request/sec_rule router\n" +
			"function f(c) end", "line 1: @stage takes phase/stage"},
		{"bad phase", "---@stage during/router\nfunction f(c) end",
			`line 1: @stage: phase "during", http:request or http:response`},
		{"stage id", "---@stage http:request/RUL\nfunction f(c) end",
			`line 1: @stage: unknown stage "RUL", one of cache, ` +
				"sec_rate_limit, waiting_room, static, sec_rule, " +
				"sec_ruleset, router"},
		{"local function", "---@stage http:request/router\n" +
			"local function f(c) end",
			"line 1: @stage must precede a global function"},
		{"method", "local M = {}\n---@stage http:request/router\n" +
			"function M.f(c) end",
			"line 2: @stage must precede a global function"},
		{"at the end", "function f(c) end\n---@stage http:response/router",
			"line 2: @stage must precede a global function"},
		{"first of the block", "-- x\n---@stage http:response/router\n" +
			"---@stage http:request/router\nlocal x = 1",
			"line 2: @stage must precede a global function"},
	}
}

func TestStagesCopy(t *testing.T) {
	c := Stages()
	c[0] = "x"

	assert.Equal(t, "CDN", Stages()[0])
}
