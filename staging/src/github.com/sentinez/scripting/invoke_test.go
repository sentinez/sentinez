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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvokeResult(t *testing.T) {
	tests := []struct {
		name string
		give string
		want Value
	}{
		{"nil", `function match(ctx) end`, Value{Kind: KindNil}},
		{"false", `function match(ctx) return false end`,
			Value{Kind: KindBool}},
		{"true", `function match(ctx) return true end`,
			Value{Kind: KindBool, Bool: true}},
		{"string", `function match(ctx) return ctx:ip() end`,
			Value{Kind: KindString, Str: "10.0.0.1"}},
		{"constant", `function match(ctx) return SZ_BLOCK end`,
			Value{Kind: KindString, Str: "block"}},
		{"table", `function match(ctx)
			return {type = SZ_REDIRECT, params = {n = 1, ok = true}}
		end`, Value{Kind: KindTable, Table: map[string]any{
			"type":   "redirect",
			"params": map[string]any{"n": float64(1), "ok": true},
		}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Compile("rule", tt.give)
			require.NoError(t, err)

			got, err := p.Invoke(context.Background(), "match", newFake())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInvokeError(t *testing.T) {
	tests := []struct {
		name string
		give string
	}{
		{"missing", `function other(ctx) end`},
		{"runtime", `function match(ctx) error("boom") end`},
		{"function result", `function match(ctx) return print end`},
		{"array result", `function match(ctx) return {1, 2} end`},
		{"set_header", `function match(ctx) ctx:set_header("a", "b") end`},
		{"set_response_header", `function match(ctx)
			ctx:set_response_header("a", "b")
		end`},
		{"respond", `function match(ctx) ctx:respond(403) end`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Compile("rule", tt.give)
			require.NoError(t, err)

			fake := newFake()
			_, err = p.Invoke(context.Background(), "match", fake)
			require.Error(t, err)
			assert.False(t, fake.answered)
			assert.Empty(t, fake.setHdr)
			assert.Empty(t, fake.respHdr)
		})
	}
}

func TestInvokeTimeout(t *testing.T) {
	p, err := Compile("rule", `function match(ctx) while true do end end`,
		WithTimeout(10*time.Millisecond))
	require.NoError(t, err)

	_, err = p.Invoke(context.Background(), "match", newFake())
	require.Error(t, err)
}

func TestHasFuncHooked(t *testing.T) {
	p, err := Compile("rule", `function match(ctx) end`)
	require.NoError(t, err)
	assert.True(t, p.HasFunc("match"))
	assert.False(t, p.HasFunc("key"))
	assert.False(t, p.Hooked())

	p, err = Compile("hook", "---@stage http.request/router\n"+
		"function h(ctx) end")
	require.NoError(t, err)
	assert.True(t, p.Hooked())
}
