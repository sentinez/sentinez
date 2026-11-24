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

package secure

import (
	"net/http"
	"testing"

	"github.com/sentinez/scripting"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory/secratelimits"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func scriptRule(t *testing.T, ns, src string) []secrules.Entry {
	t.Helper()

	store := secrules.New()
	err := store.Store(ns, []*secrulepb.SecRule{{
		Id:     "srule_script",
		Name:   "script",
		Status: typepb.Status_STATUS_ACTIVE,
		Script: src,
	}})
	require.NoError(t, err)
	t.Cleanup(func() { store.Delete(ns) })

	return store.Load(ns)
}

type scriptEvalTest struct {
	name     string
	give     string
	want     rulepb.ActionType
	wantMeta string
}

func scriptEvalTests() []scriptEvalTest {
	return []scriptEvalTest{
		{"no match", `function match(ctx) end`, _none, ""},
		{"false", `function match(ctx) return false end`, _none, ""},
		{"named", `function match(ctx) return "block" end`, _block,
			"script"},
		{"table", `function match(ctx)
			return {type = "redirect", params = {location = "/new"}}
		end`, _redirect, "script"},
		{"unknown action", `function match(ctx) return "nuke" end`,
			_none, ""},
		{"runtime error", `function match(ctx) error("boom") end`,
			_none, ""},
		{"path", `function match(ctx)
			if ctx:path() == "/x" then return "block" end
		end`, _block, "script"},
	}
}

func TestEvaluateScript(t *testing.T) {
	tests := scriptEvalTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := scriptRule(t, "script-"+tt.name, tt.give)
			ctx := newTestContext(http.MethodGet, "/x", "", nil)
			mr := &rulepb.MatchedRules{}

			got, acts := evaluate(ctx, chain, mr, nil)

			assert.Equal(t, tt.want, got.GetType())
			assert.Empty(t, acts)
			if tt.wantMeta == "" {
				assert.Empty(t, mr.GetNames())
				return
			}

			assert.Equal(t, []string{tt.wantMeta}, mr.GetNames())
		})
	}
}

func TestEvaluateScriptParams(t *testing.T) {
	chain := scriptRule(t, "script-params", `function match(ctx)
		return {type = "redirect", params = {location = "/new"}}
	end`)
	ctx := newTestContext(http.MethodGet, "/", "", nil)

	got, _ := evaluate(ctx, chain, &rulepb.MatchedRules{}, nil)

	assert.Equal(t, "/new", paramString(got, _paramLocation))
}

func TestScriptRuleRejected(t *testing.T) {
	tests := []struct {
		name string
		give *secrulepb.SecRule
	}{
		{"no match func", &secrulepb.SecRule{Script: `function x() end`}},
		{"stage hook", &secrulepb.SecRule{
			Script: "---@stage http:request/router\n" +
				"function match(ctx) end",
		}},
		{"syntax", &secrulepb.SecRule{Script: `function (`}},
		{"expr and script", &secrulepb.SecRule{
			Script: `function match(ctx) end`,
			Expr: &rulepb.Expression{
				OrCondition: []*rulepb.AndCondition{{}},
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.give.Status = typepb.Status_STATUS_ACTIVE

			err := secrules.New().Store("script-reject", []*secrulepb.SecRule{
				tt.give,
			})

			require.Error(t, err)
			assert.Empty(t, secrules.New().Load("script-reject"))
		})
	}
}

func TestScriptKey(t *testing.T) {
	tests := []struct {
		name    string
		give    scripting.Value
		want    string
		wantOk  bool
		wantErr bool
	}{
		{"nil", scripting.Value{Kind: scripting.KindNil}, "", false, false},
		{"false", scripting.Value{Kind: scripting.KindBool}, "", false, false},
		{"true", scripting.Value{Kind: scripting.KindBool, Bool: true},
			"1.1.1.1", true, false},
		{"string", scripting.Value{Kind: scripting.KindString, Str: "k"},
			"k", true, false},
		{"empty string", scripting.Value{Kind: scripting.KindString},
			"", false, false},
		{"too long", scripting.Value{
			Kind: scripting.KindString, Str: string(make([]byte, 257)),
		}, "", false, true},
		{"table", scripting.Value{Kind: scripting.KindTable},
			"", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := scriptKey(tt.give, "1.1.1.1")

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestCountedScript(t *testing.T) {
	p, err := scripting.Compile("limit", `function key(ctx)
		local k = ctx:header("X-Api-Key")
		if k == "" then return true end
		return k
	end`)
	require.NoError(t, err)

	e := secratelimits.Entry{
		Script: p, Rule: &secrulepb.SecRule{Id: "srlim_1", Name: "per key"},
	}
	ctx := newTestContext(http.MethodGet, "/", "", nil)
	run := scriptRun{ctx: ctx}
	defer run.Close()

	mr := &rulepb.MatchedRules{}
	key, ok := counted(&run, e, "9.9.9.9", mr)
	assert.True(t, ok)
	assert.Equal(t, "9.9.9.9", key)

	ctx.SetHeader([]byte("X-Api-Key"), []byte("abc"))
	key, ok = counted(&run, e, "9.9.9.9", mr)
	assert.True(t, ok)
	assert.Equal(t, "abc", key)
	assert.Equal(t, []string{"per key", "per key"}, mr.GetNames())
}
