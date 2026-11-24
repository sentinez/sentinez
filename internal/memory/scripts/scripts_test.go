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

package scripts

import (
	"testing"

	"github.com/sentinez/scripting"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	s := New()
	good := &edgepb.Script{
		Id: "a",
		Source: "---@stage http.request/router\n" +
			"function f(c) end",
	}
	off := &edgepb.Script{Id: "b", Source: "syntax error (", Disabled: true}

	require.NoError(t, s.Store("ns", []*edgepb.Script{good, off}))
	assert.Len(t, s.Load("ns"), 1, "a disabled script is not compiled")

	first := s.Load("ns")[0]
	require.NoError(t, s.Store("ns", []*edgepb.Script{good}))
	assert.Same(t, first, s.Load("ns")[0], "an unchanged script is kept")

	bad := &edgepb.Script{Id: "c", Source: "function ("}
	require.Error(t, s.Store("ns", []*edgepb.Script{bad}))
	assert.Len(t, s.Load("ns"), 1, "the previous scripts stay")

	require.NoError(t, s.Store("ns", nil))
	assert.Empty(t, s.Load("ns"))

	assert.Empty(t, (*Scripts)(nil).LoadContext(nil))
}

const (
	_matchSrc = `function match(ctx)
  if ctx:path() == "/admin" then
    return "block"
  end
end`
	_keySrc = `function key(ctx)
  if ctx:path() == "/login" then
    return true
  end
end`
	_luaSrc = `function match(ctx) return ctx:header("X") end`

	_none  = rulepb.ActionType_ACTION_TYPE_UNSPECIFIED
	_block = rulepb.ActionType_ACTION_TYPE_BLOCK
)

func TestRule(t *testing.T) {
	tests := []struct {
		name       string
		give       *secrulepb.SecRule
		wantErr    error
		wantAction rulepb.ActionType
	}{
		{"match", &secrulepb.SecRule{Script: _matchSrc}, nil, _block},
		{"no script", &secrulepb.SecRule{}, nil, _none},
		{"lua", &secrulepb.SecRule{Script: _luaSrc},
			scripting.ErrNotDeclarative, _none},
		{"expr too", &secrulepb.SecRule{Script: _matchSrc,
			Expr: &rulepb.Expression{
				OrCondition: []*rulepb.AndCondition{{}},
			}}, ErrExprAndScript, _none},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Rule(tt.give)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantAction, got.GetAction().GetType())
			assert.Empty(t, got.GetScript())
		})
	}
}

func TestRuleKeepsSource(t *testing.T) {
	give := &secrulepb.SecRule{Id: "a", Script: _matchSrc}

	got, err := Rule(give)
	require.NoError(t, err)
	assert.Len(t, got.GetExpr().GetOrCondition(), 1)
	assert.Equal(t, "a", got.GetId())
	assert.Equal(t, _matchSrc, give.GetScript(), "r is not modified")
}

func TestKeyRule(t *testing.T) {
	got, err := KeyRule(&secrulepb.SecRule{Script: _keySrc})
	require.NoError(t, err)
	assert.Len(t, got.GetExpr().GetOrCondition(), 1)
	assert.Empty(t, got.GetScript())

	_, err = KeyRule(&secrulepb.SecRule{
		Script: `function key(ctx) return ctx:header("K") end`,
	})
	require.ErrorIs(t, err, scripting.ErrNotDeclarative)
}
