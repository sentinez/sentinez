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
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// callHook compiles a script whose @types before ROU hook runs body, and
// calls it.
func callHook(t *testing.T, body string) *fakeExchange {
	t.Helper()

	p, err := Compile("t.lua",
		"---@module ROU\n---@types before\nfunction f(ctx)\n"+body+"\nend")
	require.NoError(t, err)

	ex := newFake()
	_, err = p.Call(context.Background(), Before, "ROU", ex)
	require.NoError(t, err)

	return ex
}

func TestLog(t *testing.T) {
	tests := []struct {
		name string
		give string
		want []string
	}{
		{"info", `log.info("blocked", ctx:path())`,
			[]string{"info t.lua: blocked\t/api"}},
		{"debug", `log.debug("x-debug:", ctx:header("X-Debug"))`,
			[]string{"debug t.lua: x-debug:\t1"}},
		{"tostring", `log.info(1, nil, true, 2.5)`,
			[]string{"info t.lua: 1\tnil\ttrue\t2.5"}},
		{"__tostring", `log.info(setmetatable({}, {__tostring = ` +
			`function() return "obj" end}))`,
			[]string{"info t.lua: obj"}},
		{"no argument", `log.debug()`, []string{"debug t.lua: "}},
		{"colon", `log:info("a", ctx:method())`,
			[]string{"info t.lua: a\tGET"}},
		{"colon without argument", `log:debug()`,
			[]string{"debug t.lua: "}},
		{"another table first", `log.info(setmetatable({}, ` +
			`{__tostring = function() return "other" end}), 1)`,
			[]string{"info t.lua: other\t1"}},
		{"several lines", `log.info("a") log:debug("b")`,
			[]string{"info t.lua: a", "debug t.lua: b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, callHook(t, tt.give).logs)
		})
	}
}

func TestLogLimits(t *testing.T) {
	long := callHook(t, `log.info(string.rep("é", 2000))`)
	require.Len(t, long.logs, 1)
	assert.LessOrEqual(t, len(long.logs[0]), len("info t.lua: ")+_maxLogLen)
	assert.True(t, strings.HasSuffix(long.logs[0], "é"), "cut on a rune")

	many := callHook(t, `for i = 1, 100 do log:info(i) end`)
	assert.Len(t, many.logs, _maxLogLines)
}

func TestLogInScriptBodyIsDropped(t *testing.T) {
	p, err := Compile("t.lua", `
log.info("loaded")
log:debug("loaded")
---@module ROU
---@types before
function f(ctx) end
`)
	require.NoError(t, err)

	ex := newFake()
	_, err = p.Call(context.Background(), Before, "ROU", ex)
	require.NoError(t, err)
	assert.Empty(t, ex.logs)
}

func TestLogIsNotAFunction(t *testing.T) {
	p, err := Compile("t.lua",
		"---@module ROU\n---@types before\nfunction f(ctx) "+
			`log("info", "x")`+" end")
	require.NoError(t, err)

	_, err = p.Call(context.Background(), Before, "ROU", newFake())
	assert.ErrorContains(t, err, "attempt to call a non-function object")
}
