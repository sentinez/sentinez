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

package rulescript

import (
	"testing"

	"github.com/sentinez/scripting"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestFromScriptFlat(t *testing.T) {
	src := `function match(ctx)
  if ctx:path() == "/old" or ctx:header("X-Old") ~= "" then
    return {type = "redirect", params = {location = "/new", status = 301}}
  end
end`

	expr, action, err := FromScript(src)
	require.NoError(t, err)

	params, err := structpb.NewStruct(map[string]any{
		"location": "/new", "status": 301,
	})
	require.NoError(t, err)

	wantExpr := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{
		{Rules: []*rulepb.Rule{cond(_path, "", _eq, str("/old"))}},
		{Rules: []*rulepb.Rule{cond(_header, "X-Old", _ne, str(""))}},
	}}
	wantAction := &rulepb.Action{
		Type: rulepb.ActionType_ACTION_TYPE_REDIRECT, Params: params,
	}

	assert.True(t, proto.Equal(wantExpr, expr), "%v", expr)
	assert.True(t, proto.Equal(wantAction, action), "%v", action)

	// and back to the same script
	again, err := ToScript(expr, action)
	require.NoError(t, err)

	expr2, action2, err := FromScript(again)
	require.NoError(t, err)
	assert.True(t, proto.Equal(expr, expr2))
	assert.True(t, proto.Equal(action, action2))
}

func TestKeyScriptRoundTrip(t *testing.T) {
	expr := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{
		{Rules: []*rulepb.Rule{cond(_path, "",
			rulepb.Operator_OPERATOR_PREFIX, str("/"))}},
	}}

	src, err := ToKeyScript(expr)
	require.NoError(t, err)

	got, err := FromKeyScript(src)
	require.NoError(t, err)
	assert.True(t, proto.Equal(expr, got), "%v", got)
}

func TestToScriptUnsupported(t *testing.T) {
	block := &rulepb.Action{Type: rulepb.ActionType_ACTION_TYPE_BLOCK}
	tests := map[string]*rulepb.Rule{
		"regex": cond(_path, "", rulepb.Operator_OPERATOR_MATCHES,
			str("^/a")),
		"cidr": cond(_ip, "", _eq, str("10.0.0.0/8")),
		"body": cond(rulepb.FieldSource_FIELD_SOURCE_BODY, "", _eq,
			str("")),
		"presence": cond(_header, "", _in, list(t, "X-A")),
		"number":   cond(_path, "", _eq, structpb.NewNumberValue(1)),
		"path in":  cond(_path, "", _in, list(t, "/a")),
	}

	for name, r := range tests {
		t.Run(name, func(t *testing.T) {
			expr := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{
				{Rules: []*rulepb.Rule{r}},
			}}

			_, err := ToScript(expr, block)
			require.ErrorIs(t, err, scripting.ErrNotDeclarative)
		})
	}
}

func TestToScriptRateLimitAction(t *testing.T) {
	_, err := ToScript(nil, &rulepb.Action{
		Type: rulepb.ActionType_ACTION_TYPE_RATE_LIMIT,
	})

	require.ErrorIs(t, err, scripting.ErrNotDeclarative)
}
