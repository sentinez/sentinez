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
---@module RUL 'RST'
---@param ctx sz.Context
---@types before
function whatever_name(ctx) end

---@module "ROU"
---@types after
function on_router(ctx) end

function before_cache(ctx) end -- a name alone declares nothing

---@module STC
---@types before after
function both(ctx) end

---@module 'socket'
local socket = {} -- no stage, no @types: left alone
`

func TestCompileHooks(t *testing.T) {
	p, err := Compile("t.lua", _hooksScript)
	require.NoError(t, err)

	tests := []struct {
		givePhase Phase
		giveStage string
		want      bool
	}{
		{Before, "RUL", true},
		{Before, "RST", true},
		{After, "ROU", true},
		{Before, "STC", true},
		{After, "STC", true},
		{Before, "ROU", false},
		{After, "RUL", false},
		{Before, "CDN", false},
	}

	for _, tt := range tests {
		name := tt.givePhase.String() + " " + tt.giveStage
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, p.Has(tt.givePhase, tt.giveStage))
		})
	}
}

func TestCompileInvalidAnnotation(t *testing.T) {
	tests := []struct {
		name    string
		give    string
		wantErr string
	}{
		{"empty types", "---@module RST\n---@types\nfunction f(c) end",
			"line 2: @types names nothing"},
		{"empty module", "---@module ''\n---@types after\nfunction f(c) end",
			"line 1: @module names nothing"},
		{"bad types", "---@module RST\n---@types during\nfunction f(c) end",
			`line 2: @types "during": before or after`},
		{"no types", "---@module RST\nfunction f(c) end",
			"line 1: @module RST needs @types before or after"},
		{"no class", "---@types after\nfunction f(c) end",
			"line 1: @types needs a @module stage"},
		{"unknown stage", "---@module RST proxy\n---@types after\n" +
			"function f(c) end", `line 1: @module: unknown stage "proxy"`},
		{"local function", "---@module RST\n---@types before\n" +
			"local function f(c) end",
			"line 1: a hook must be a global function"},
		{"method", "local M = {}\n---@module RST\n---@types before\n" +
			"function M.f(c) end", "line 2: a hook must be a global function"},
		{"at the end", "function f(c) end\n---@module RST\n---@types after",
			"line 2: a hook must be a global function"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compile("t.lua", tt.give)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestStagesCopy(t *testing.T) {
	c := Stages()
	c[0] = "x"

	assert.Equal(t, "CDN", Stages()[0])
}
