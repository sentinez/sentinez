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

package corers

import (
	"strings"
	"testing"

	"github.com/corazawaf/coraza/v3"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// _attackRule stands in for a CRS rule, it blocks any "attack" argument.
const _attackRule = `SecRule ARGS "@contains attack" ` +
	`"id:942100,phase:2,deny,status:403,tag:'attack-sqli'"` + "\n"

// The request every test case sends, carrying an attack.
const (
	_giveMethod = "POST"
	_givePath   = "/api/search"
	_giveQuery  = "q=attack&lang=vi"
	_giveHost   = "api.example.com"
	_giveClient = "10.0.0.10"
	_giveAgent  = "curl/8.0.1"
)

func cond(
	source rulepb.FieldSource, key string, op rulepb.Operator, value any,
) *rulepb.Condition {
	v, err := structpb.NewValue(value)
	if err != nil {
		panic(err)
	}

	return &rulepb.Condition{
		Source: source, Key: key, Operator: op, Value: v,
	}
}

// and a conjunction of conditions, optionally and one of the nested.
func and(
	nested []*rulepb.AndCondition, conds ...*rulepb.Condition,
) *rulepb.AndCondition {
	rules := make([]*rulepb.Rule, 0, len(conds))
	for _, c := range conds {
		rules = append(rules, &rulepb.Rule{Condition: c})
	}

	return &rulepb.AndCondition{Rules: rules, OrCondition: nested}
}

func expr(or ...*rulepb.AndCondition) *rulepb.Expression {
	return &rulepb.Expression{OrCondition: or}
}

// exprOf an expression of a single condition.
func exprOf(
	source rulepb.FieldSource, key string, op rulepb.Operator, value any,
) *rulepb.Expression {
	return expr(and(nil, cond(source, key, op, value)))
}

func newSetting(e *corerulesetpb.Exclusion) *corerulesetpb.CoreRuleset {
	return &corerulesetpb.CoreRuleset{
		Mode:       corerulesetpb.EngineMode_ENGINE_MODE_ON,
		Status:     typepb.Status_STATUS_ACTIVE,
		Exclusions: []*corerulesetpb.Exclusion{e},
	}
}

// blocked sends the request through a WAF built from the exclusion and
// the attack rule.
func blocked(t *testing.T, exclusion *corerulesetpb.Exclusion) bool {
	t.Helper()

	return blockedBy(t, newSetting(exclusion))
}

func blockedBy(t *testing.T, setting *corerulesetpb.CoreRuleset) bool {
	t.Helper()

	require.NoError(t, validateSetting(setting))

	var sb strings.Builder
	sb.WriteString("SecRuleEngine On\n")
	writeScopedRules(&sb, setting)
	sb.WriteString(_attackRule)
	sb.WriteString(globalDirectives(
		setting, func(uint32) bool { return true },
	))

	waf, err := coraza.NewWAF(
		coraza.NewWAFConfig().WithDirectives(sb.String()),
	)
	require.NoError(t, err, sb.String())

	tx := waf.NewTransaction()
	t.Cleanup(func() { assert.NoError(t, tx.Close()) })

	tx.ProcessConnection(_giveClient, 52341, "", 0)
	tx.ProcessURI(_givePath+"?"+_giveQuery, _giveMethod, "HTTP/1.1")
	tx.AddRequestHeader("Host", _giveHost)
	tx.AddRequestHeader("User-Agent", _giveAgent)
	tx.SetServerName(_giveHost)
	tx.ProcessRequestHeaders()
	_, err = tx.ProcessRequestBody()
	require.NoError(t, err)

	return tx.IsInterrupted()
}

// nolint:funlen
func TestExclusionExpr(t *testing.T) {
	const (
		path   = rulepb.FieldSource_FIELD_SOURCE_PATH
		host   = rulepb.FieldSource_FIELD_SOURCE_HOST
		method = rulepb.FieldSource_FIELD_SOURCE_METHOD
		ip     = rulepb.FieldSource_FIELD_SOURCE_IP
		header = rulepb.FieldSource_FIELD_SOURCE_HEADER
		query  = rulepb.FieldSource_FIELD_SOURCE_QUERY

		eq       = rulepb.Operator_OPERATOR_EQ
		ne       = rulepb.Operator_OPERATOR_NE
		contains = rulepb.Operator_OPERATOR_CONTAINS
		matches  = rulepb.Operator_OPERATOR_MATCHES
		prefix   = rulepb.Operator_OPERATOR_PREFIX
		suffix   = rulepb.Operator_OPERATOR_SUFFIX
		in       = rulepb.Operator_OPERATOR_IN
		notIn    = rulepb.Operator_OPERATOR_NOT_IN
	)

	list := func(values ...any) []any { return values }

	tests := []struct {
		name        string
		give        *rulepb.Expression
		wantBlocked bool
	}{
		{name: "path eq", give: exprOf(path, "", eq, "/api/search")},
		{
			name: "path eq, other", give: exprOf(path, "", eq, "/api"),
			wantBlocked: true,
		},
		{name: "path ne", give: exprOf(path, "", ne, "/web")},
		{name: "path prefix", give: exprOf(path, "", prefix, "/api/")},
		{name: "path suffix", give: exprOf(path, "", suffix, "/search")},
		{name: "path contains", give: exprOf(path, "", contains, "i/s")},
		{
			name: "path matches",
			give: exprOf(path, "", matches, `^/api/\w+$`),
		},
		{
			name:        "path matches, other",
			give:        exprOf(path, "", matches, `^/api/\d+$`),
			wantBlocked: true,
		},
		{name: "host eq", give: exprOf(host, "", eq, "api.example.com")},
		{
			name: "host eq, other", give: exprOf(host, "", eq, "example.com"),
			wantBlocked: true,
		},
		{name: "method eq", give: exprOf(method, "", eq, "POST")},
		{
			name: "method in",
			give: exprOf(method, "", in, list("PUT", "POST")),
		},
		{
			name:        "method in, substring",
			give:        exprOf(method, "", in, list("POSTS", "PO")),
			wantBlocked: true,
		},
		{
			name:        "method not in",
			give:        exprOf(method, "", notIn, list("PUT", "POST")),
			wantBlocked: true,
		},
		{name: "ip eq", give: exprOf(ip, "", eq, "10.0.0.10")},
		{name: "ip eq, cidr", give: exprOf(ip, "", eq, "10.0.0.0/24")},
		{
			name: "ip ne", give: exprOf(ip, "", ne, "10.0.0.0/24"),
			wantBlocked: true,
		},
		{
			name: "ip in",
			give: exprOf(ip, "", in, list("192.168.0.0/16", "10.0.0.10")),
		},
		{
			name:        "ip not in",
			give:        exprOf(ip, "", notIn, list("10.0.0.0/8")),
			wantBlocked: true,
		},
		{
			name: "header value",
			give: exprOf(header, "User-Agent", prefix, "curl/"),
		},
		{
			name:        "header value, other",
			give:        exprOf(header, "User-Agent", prefix, "wget/"),
			wantBlocked: true,
		},
		{
			name: "header value in",
			give: exprOf(header, "User-Agent", in, list("a", "curl/8.0.1")),
		},
		{name: "header present", give: exprOf(header, "", eq, "User-Agent")},
		{
			name:        "header present, missing",
			give:        exprOf(header, "", eq, "X-Internal"),
			wantBlocked: true,
		},
		{name: "header absent", give: exprOf(header, "", ne, "X-Internal")},
		{
			name: "headers present",
			give: exprOf(header, "", in, list("Host", "User-Agent")),
		},
		{
			name:        "headers present, one missing",
			give:        exprOf(header, "", in, list("Host", "X-Internal")),
			wantBlocked: true,
		},
		{
			name: "headers absent",
			give: exprOf(header, "", notIn, list("X-A", "X-B")),
		},
		{name: "query value", give: exprOf(query, "lang", eq, "vi")},
		{
			name: "query value, other", give: exprOf(query, "lang", eq, "en"),
			wantBlocked: true,
		},
		{name: "query present", give: exprOf(query, "", eq, "lang")},
		{
			name: "and",
			give: expr(and(nil,
				cond(method, "", eq, "POST"), cond(ip, "", eq, "10.0.0.10"),
			)),
		},
		{
			name: "and, one not matched",
			give: expr(and(nil,
				cond(method, "", eq, "POST"), cond(ip, "", eq, "10.0.0.11"),
			)),
			wantBlocked: true,
		},
		{
			name: "or, second matched",
			give: expr(
				and(nil, cond(method, "", eq, "GET")),
				and(nil, cond(ip, "", eq, "10.0.0.10")),
			),
		},
		{
			name: "or, none matched",
			give: expr(
				and(nil, cond(method, "", eq, "GET")),
				and(nil, cond(ip, "", eq, "10.0.0.11")),
			),
			wantBlocked: true,
		},
		{
			name: "nested or matched",
			give: expr(and(
				[]*rulepb.AndCondition{
					and(nil, cond(ip, "", eq, "10.0.0.11")),
					and(nil, cond(query, "lang", eq, "vi")),
				},
				cond(method, "", eq, "POST"),
			)),
		},
		{
			name: "nested or not matched",
			give: expr(and(
				[]*rulepb.AndCondition{
					and(nil, cond(ip, "", eq, "10.0.0.11")),
					and(nil, cond(query, "lang", eq, "en")),
				},
				cond(method, "", eq, "POST"),
			)),
			wantBlocked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blocked(t, &corerulesetpb.Exclusion{
				Tags: []string{"attack-sqli"},
				Expr: tt.give,
			})

			assert.Equal(t, tt.wantBlocked, got)
		})
	}
}

func TestExclusionScope(t *testing.T) {
	tags := []string{"attack-sqli"}
	targets := []*corerulesetpb.ExclusionTarget{
		{RuleId: 942100, Variable: "ARGS:q"},
	}
	post := exprOf(
		rulepb.FieldSource_FIELD_SOURCE_METHOD, "",
		rulepb.Operator_OPERATOR_EQ, "POST",
	)

	tests := []struct {
		name        string
		give        *corerulesetpb.Exclusion
		wantBlocked bool
	}{
		{
			name:        "nothing excluded",
			give:        &corerulesetpb.Exclusion{},
			wantBlocked: true,
		},
		{name: "global", give: &corerulesetpb.Exclusion{Tags: tags}},
		{
			name: "target with expr",
			give: &corerulesetpb.Exclusion{Targets: targets, Expr: post},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantBlocked, blocked(t, tt.give))
		})
	}
}

func TestOverrideExpr(t *testing.T) {
	method := func(m string) *rulepb.Expression {
		return exprOf(
			rulepb.FieldSource_FIELD_SOURCE_METHOD, "",
			rulepb.Operator_OPERATOR_EQ, m,
		)
	}

	tests := []struct {
		name        string
		give        *rulepb.Expression
		wantBlocked bool
	}{
		{name: "global"},
		{name: "expr matched", give: method("POST")},
		{name: "expr not matched", give: method("GET"), wantBlocked: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blockedBy(t, &corerulesetpb.CoreRuleset{
				Overrides: []*corerulesetpb.RuleOverride{{
					Id:    942100,
					State: corerulesetpb.RuleState_RULE_STATE_DISABLED,
					Expr:  tt.give,
				}},
			})

			assert.Equal(t, tt.wantBlocked, got)
		})
	}
}

// nolint:funlen
func TestOverrideExprInvalid(t *testing.T) {
	post := exprOf(
		rulepb.FieldSource_FIELD_SOURCE_METHOD, "",
		rulepb.Operator_OPERATOR_EQ, "POST",
	)
	body := exprOf(
		rulepb.FieldSource_FIELD_SOURCE_BODY, "",
		rulepb.Operator_OPERATOR_EQ, "a",
	)

	tests := []struct {
		name      string
		give      *rulepb.Expression
		giveState corerulesetpb.RuleState
	}{
		{
			name:      "detection only scoped",
			give:      post,
			giveState: corerulesetpb.RuleState_RULE_STATE_DETECTION_ONLY,
		},
		{
			name:      "enabled scoped",
			give:      post,
			giveState: corerulesetpb.RuleState_RULE_STATE_ENABLED,
		},
		{
			name:      "source",
			give:      body,
			giveState: corerulesetpb.RuleState_RULE_STATE_DISABLED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSetting(&corerulesetpb.CoreRuleset{
				Overrides: []*corerulesetpb.RuleOverride{{
					Id: 942100, State: tt.giveState, Expr: tt.give,
				}},
			})

			require.ErrorIs(t, err, ErrInvalidSetting)
		})
	}
}

// nolint:funlen
func TestExclusionExprInvalid(t *testing.T) {
	const (
		path = rulepb.FieldSource_FIELD_SOURCE_PATH
		eq   = rulepb.Operator_OPERATOR_EQ
	)

	tests := []struct {
		name string
		give *rulepb.Expression
	}{
		{
			name: "body",
			give: exprOf(rulepb.FieldSource_FIELD_SOURCE_BODY, "", eq, "a"),
		},
		{
			name: "ja4",
			give: exprOf(rulepb.FieldSource_FIELD_SOURCE_JA4, "", eq, "a"),
		},
		{
			name: "tls",
			give: exprOf(rulepb.FieldSource_FIELD_SOURCE_TLS, "", eq, "true"),
		},
		{name: "no condition", give: expr(and(nil, nil))},
		{name: "empty value", give: exprOf(path, "", eq, "")},
		{name: "quote", give: exprOf(path, "", eq, `/a" "id:1,deny`)},
		{name: "macro", give: exprOf(path, "", eq, "%{tx.a}")},
		{name: "trailing backslash", give: exprOf(path, "", eq, `/a\`)},
		{name: "new line", give: exprOf(path, "", eq, "/a\nSecRuleEngine Off")},
		{
			name: "operator",
			give: exprOf(path, "", rulepb.Operator_OPERATOR_GT, "/a"),
		},
		{
			name: "regex",
			give: exprOf(path, "", rulepb.Operator_OPERATOR_MATCHES, "("),
		},
		{
			name: "address",
			give: exprOf(rulepb.FieldSource_FIELD_SOURCE_IP, "", eq, "10.0.0"),
		},
		{
			name: "key",
			give: exprOf(
				rulepb.FieldSource_FIELD_SOURCE_HEADER, `a"b`, eq, "a",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSetting(newSetting(&corerulesetpb.Exclusion{
				Tags: []string{"attack-sqli"}, Expr: tt.give,
			}))

			require.ErrorIs(t, err, ErrInvalidSetting)
		})
	}
}
