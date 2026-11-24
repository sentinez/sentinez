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
	"strings"
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
		"ip contains": cond(_ip, "", rulepb.Operator_OPERATOR_CONTAINS,
			str("10.")),
		"tls value": cond(_tls, "", _eq, str("yes")),
		"presence op": cond(_header, "", rulepb.Operator_OPERATOR_PREFIX,
			str("X-A")),
		"presence empty": cond(_query, "query", _in, list(t, "")),
		"bad regex": cond(_path, "", rulepb.Operator_OPERATOR_MATCHES,
			str("(")),
		"number":  cond(_path, "", _eq, structpb.NewNumberValue(1)),
		"path in": cond(_path, "", _in, list(t, "/a")),
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

// TestReadsRoundTrip the conditions of the regex, CIDR, body, JA4, TLS
// and presence come back the same from their script.
func TestReadsRoundTrip(t *testing.T) {
	rules := []*rulepb.Rule{
		cond(_path, "", rulepb.Operator_OPERATOR_MATCHES, str("^/wp-")),
		cond(_ip, "", _eq, str("10.0.0.0/8")),
		cond(_ip, "", _ne, str("192.168.0.0/16")),
		cond(rulepb.FieldSource_FIELD_SOURCE_BODY, "",
			rulepb.Operator_OPERATOR_CONTAINS, str("<script")),
		cond(rulepb.FieldSource_FIELD_SOURCE_JA4, "", _eq, str("t13d")),
		cond(_tls, "", _ne, str("true")),
		cond(_header, "", _eq, str("X-Debug")),
		cond(_query, "", _ne, str("debug")),
	}
	expr := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{
		{Rules: rules},
	}}
	block := &rulepb.Action{Type: rulepb.ActionType_ACTION_TYPE_BLOCK}

	src, err := ToScript(expr, block)
	require.NoError(t, err)

	got, _, err := FromScript(src)
	require.NoError(t, err, src)
	assert.True(t, proto.Equal(expr, got), "%s\n%v", src, got)
}

// orOf the or group of the conditions: an and condition each.
func orOf(rules ...*rulepb.Rule) []*rulepb.AndCondition {
	ors := make([]*rulepb.AndCondition, 0, len(rules))
	for _, r := range rules {
		ors = append(ors, &rulepb.AndCondition{Rules: []*rulepb.Rule{r}})
	}

	return ors
}

// nestedExpr ip != 1 and (path = /a or path = /b) and (header X or header
// Y) and not (query q = 1 and (host = h or host = i)), in the form
// ToScript reads it back to.
func nestedExpr() *rulepb.Expression {
	host := func(v string) *rulepb.Rule {
		return cond(rulepb.FieldSource_FIELD_SOURCE_HOST, "", _eq, str(v))
	}

	return &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{
		Rules: []*rulepb.Rule{cond(_ip, "", _ne, str("1"))},
		OrCondition: orOf(cond(_path, "", _eq, str("/a")),
			cond(_path, "", _eq, str("/b"))),
		AndCondition: []*rulepb.AndCondition{
			{OrCondition: orOf(cond(_header, "X", _eq, str("1")),
				cond(_header, "Y", _eq, str("1")))},
			{OrCondition: []*rulepb.AndCondition{
				{Rules: []*rulepb.Rule{host("h"), host("i")}},
				{Rules: []*rulepb.Rule{cond(_query, "q", _ne, str("1"))}},
			}},
		},
	}, {
		Rules: []*rulepb.Rule{cond(_path, "", _eq, str("/z"))},
	}}}
}

// TestNestedRoundTrip an expression of several or groups in an and group
// comes back from its script as it was, with no branch repeated.
func TestNestedRoundTrip(t *testing.T) {
	expr := nestedExpr()
	block := &rulepb.Action{Type: rulepb.ActionType_ACTION_TYPE_BLOCK}

	src, err := ToScript(expr, block)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(src, `ctx:path() == "/a"`), src)

	got, _, err := FromScript(src)
	require.NoError(t, err, src)
	assert.True(t, proto.Equal(expr, got), "%s\n%v", src, got)

	key, err := ToKeyScript(expr)
	require.NoError(t, err)

	gotKey, err := FromKeyScript(key)
	require.NoError(t, err, key)
	assert.True(t, proto.Equal(expr, gotKey), "%s\n%v", key, gotKey)
}

func TestToConditionNested(t *testing.T) {
	got, err := ToCondition(&rulepb.Expression{
		OrCondition: []*rulepb.AndCondition{{
			Rules: []*rulepb.Rule{cond(_ip, "", _ne, str("1"))},
			OrCondition: orOf(cond(_path, "", _eq, str("/a")),
				cond(_path, "", _eq, str("/b"))),
			AndCondition: []*rulepb.AndCondition{{
				OrCondition: orOf(cond(_header, "X", _eq, str("1")),
					cond(_header, "Y", _eq, str("1"))),
			}},
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, `ctx:ip() ~= "1" and (ctx:path() == "/a" or `+
		`ctx:path() == "/b") and (ctx:header("X") == "1" or `+
		`ctx:header("Y") == "1")`, got)
}

func TestToConditionLists(t *testing.T) {
	tests := []struct {
		name string
		give *rulepb.Rule
		want string
	}{
		{"ip in", cond(_ip, "", _in, list(t, "10.0.0.0/8", "1.1.1.1")),
			`ctx:ip_in("10.0.0.0/8") or ctx:ip() == "1.1.1.1"`},
		{"ip not in", cond(_ip, "", _notIn, list(t, "::1/128", "1.1.1.1")),
			`not ctx:ip_in("::1/128") and ctx:ip() ~= "1.1.1.1"`},
		{"headers in", cond(_header, "header", _in, list(t, "A", "B")),
			`ctx:has_header("A") and ctx:has_header("B")`},
		{"queries not in", cond(_query, "", _notIn, list(t, "a", "b")),
			`not ctx:has_query("a") and not ctx:has_query("b")`},
		{"tls", cond(_tls, "", _eq, str("false")), `not ctx:tls()`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToCondition(&rulepb.Expression{
				OrCondition: []*rulepb.AndCondition{
					{Rules: []*rulepb.Rule{tt.give}},
				},
			})
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToScriptRateLimitAction(t *testing.T) {
	_, err := ToScript(nil, &rulepb.Action{
		Type: rulepb.ActionType_ACTION_TYPE_RATE_LIMIT,
	})

	require.ErrorIs(t, err, scripting.ErrNotDeclarative)
}
