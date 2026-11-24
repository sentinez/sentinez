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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szrule"
	"github.com/sentinez/core/szrule/rulescript"
	"github.com/sentinez/scripting"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/sentinez/sentinez/pkg/pools/exchange"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func cond(src rulepb.FieldSource, key string, op rulepb.Operator,
	v *structpb.Value) *rulepb.Rule {
	return &rulepb.Rule{Condition: &rulepb.Condition{
		Source: src, Key: key, Operator: op, Value: v,
	}}
}

func str(s string) *structpb.Value { return structpb.NewStringValue(s) }

func list(t *testing.T, s ...any) *structpb.Value {
	t.Helper()

	v, err := structpb.NewList(s)
	require.NoError(t, err)

	return structpb.NewListValue(v)
}

const (
	_path   = rulepb.FieldSource_FIELD_SOURCE_PATH
	_method = rulepb.FieldSource_FIELD_SOURCE_METHOD
	_ip     = rulepb.FieldSource_FIELD_SOURCE_IP
	_header = rulepb.FieldSource_FIELD_SOURCE_HEADER
	_query  = rulepb.FieldSource_FIELD_SOURCE_QUERY

	_opEq     = rulepb.Operator_OPERATOR_EQ
	_opNe     = rulepb.Operator_OPERATOR_NE
	_opGt     = rulepb.Operator_OPERATOR_GT
	_opIn     = rulepb.Operator_OPERATOR_IN
	_opNotIn  = rulepb.Operator_OPERATOR_NOT_IN
	_opPrefix = rulepb.Operator_OPERATOR_PREFIX
)

// exprs the expressions of the tests: the ones of proxy.yaml, nested or
// conditions and lists.
func exprs(t *testing.T) map[string]*rulepb.Expression {
	return map[string]*rulepb.Expression{
		"proxy.yaml": {OrCondition: []*rulepb.AndCondition{
			{Rules: []*rulepb.Rule{
				cond(_path, "", _opEq, str("/block")),
				cond(_method, "", _opNe, str("POST")),
			}},
			{Rules: []*rulepb.Rule{cond(_ip, "", _opEq, str("1.1.1.1"))}},
		}},
		"nested": {OrCondition: []*rulepb.AndCondition{{
			Rules: []*rulepb.Rule{cond(_path, "",
				_opPrefix, str("/api"))},
			OrCondition: []*rulepb.AndCondition{
				{Rules: []*rulepb.Rule{cond(_header, "X-Debug", _opEq,
					str("1"))}},
				{Rules: []*rulepb.Rule{cond(_query, "n",
					_opGt, str("5"))}},
			},
		}}},
		"lists": {OrCondition: []*rulepb.AndCondition{{
			Rules: []*rulepb.Rule{
				cond(_method, "", _opIn, list(t, "GET", "HEAD")),
				cond(_header, "X-Tier", _opNotIn, list(t, "gold", "vip")),
			},
		}}},
		"all": nil,
	}
}

// requests the requests the expressions are evaluated on.
func requests() []*http.Request {
	var reqs []*http.Request
	for _, target := range []string{"/block", "/api/x?n=9", "/api?n=1",
		"/other"} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			for _, debug := range []string{"", "1"} {
				r := httptest.NewRequest(method, "https://a.example.com"+
					target, nil)
				r.Header.Set("X-Debug", debug)
				reqs = append(reqs, r)
			}
		}
	}

	return reqs
}

// TestToScriptMatchesExpr the script of an expr matches the requests the
// expr does, and its expr read back too.
func TestToScriptMatchesExpr(t *testing.T) {
	action := &rulepb.Action{Type: rulepb.ActionType_ACTION_TYPE_BLOCK}

	for name, expr := range exprs(t) {
		t.Run(name, func(t *testing.T) {
			src, err := rulescript.ToScript(expr, action)
			require.NoError(t, err)

			p, err := scripting.Compile(name, src)
			require.NoError(t, err, src)

			back, gotAction, err := rulescript.FromScript(src)
			require.NoError(t, err, src)
			assert.True(t, proto.Equal(action, gotAction))

			for _, r := range requests() {
				ctx := httpx.NewContext(r, httptest.NewRecorder())
				want := szrule.NewEval("a", expr)(ctx, nil)

				assert.Equal(t, want, runScript(t, p, ctx), "%s %s\n%s",
					r.Method, r.URL, src)
				assert.Equal(t, want, szrule.NewEval("b", back)(ctx, nil))
			}
		})
	}
}

func runScript(t *testing.T, p *scripting.Program, ctx szhttp.Context,
) bool {
	t.Helper()

	ex := exchange.Acquire(ctx)
	defer exchange.Release(ex)

	v, err := p.Invoke(context.Background(), scripting.MatchFunc, ex)
	require.NoError(t, err)

	return v.Kind == scripting.KindString || v.Kind == scripting.KindTable
}
