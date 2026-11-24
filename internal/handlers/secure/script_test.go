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

	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
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
		// evaluated as the expr it converts to, matched by its conditions
		{"named", `function match(ctx) return "block" end`, _block,
			"true"},
		{"table", `function match(ctx)
			return {type = "redirect", params = {location = "/new"}}
		end`, _redirect, "true"},
		{"path", `function match(ctx)
			if ctx:path() == "/x" then return "block" end
		end`, _block, `ctx:path() == "/x"`},
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

// TestScriptRuleRejected a script that does not convert to an expr fails
// the load: no Lua runs at request time. scripts.Rule tests the rule with
// an expr too.
func TestScriptRuleRejected(t *testing.T) {
	tests := []struct {
		name string
		give string
	}{
		{"no match func", `function x() end`},
		{"stage hook", "---@stage http.request/router\n" +
			"function match(ctx) end"},
		{"syntax", `function (`},
		{"no branch", `function match(ctx) end`},
		{"unknown action", `function match(ctx) return "nuke" end`},
		{"not declarative", `function match(ctx)
			local a = "block"
			return a
		end`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := secrules.New().Store("script-reject", []*secrulepb.SecRule{{
				Status: typepb.Status_STATUS_ACTIVE, Script: tt.give,
			}})

			require.Error(t, err)
			assert.Empty(t, secrules.New().Load("script-reject"))
		})
	}
}
