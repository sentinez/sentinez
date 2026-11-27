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

package secure

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sentinez/core/szcrypto/pow"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/core/szlimiter"
	"github.com/sentinez/core/szrule"
	"github.com/sentinez/pkg/szlog"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory/secratelimits"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/sentinez/sentinez/internal/memory/secrulesets"
	"github.com/sentinez/sentinez/internal/memory/settings"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// BenchmarkSecLimiterAllowed a request counted by a limiter of its
// namespace, under the limit.
func BenchmarkSecLimiterAllowed(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	const namespace = "bench.example.com"

	expr := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{}}}
	store := secratelimits.New()
	store.Store(namespace, []secratelimits.Entry{{
		Eval:    szrule.NewEval("srlim_bench", expr),
		Rule:    &secrulepb.SecRule{Id: "srlim_bench", Expr: expr},
		Limiter: szlimiter.NewRateLimiter(time.Second, time.Second, 1<<62),
	}})

	ctx := newTestContext(http.MethodGet, "/", "", nil)
	ctx.X().Namespace = namespace
	next := chains.HandlerFunc(func(szhttp.Context) error { return nil })
	plugin := &SecLimiter{store: store}
	b.ReportAllocs()

	for b.Loop() {
		_ = plugin.Handle(ctx, next)
	}
}

// BenchmarkSecLimiterExceeded a request over the limit: the event is
// built and logged, and the 429 answered.
func BenchmarkSecLimiterExceeded(b *testing.B) {
	const namespace = "bench-exceeded.example.com"

	expr := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{}}}
	store := secratelimits.New()
	store.Store(namespace, []secratelimits.Entry{{
		Eval:    szrule.NewEval("srlim_bench", expr),
		Rule:    &secrulepb.SecRule{Id: "srlim_bench", Expr: expr},
		Limiter: szlimiter.NewRateLimiter(time.Hour, time.Hour, 0),
	}})

	ctx := newTestContext(http.MethodGet, "/", "", nil)
	ctx.X().Namespace = namespace
	plugin := &SecLimiter{store: store, log: benchLog(
		typepb.LogType_LOG_TYPE_SECRATELIMIT)}
	b.ReportAllocs()

	for b.Loop() {
		_ = plugin.Handle(ctx, benchNext())
	}
}

// BenchmarkSecLimiterNoRules a request of a namespace without limiters.
func BenchmarkSecLimiterNoRules(b *testing.B) {
	ctx := newTestContext(http.MethodGet, "/", "", nil)
	ctx.X().Namespace = "bench-none.example.com"
	plugin := &SecLimiter{store: secratelimits.New()}
	b.ReportAllocs()

	for b.Loop() {
		_ = plugin.Handle(ctx, benchNext())
	}
}

// BenchmarkSecRuleNoRules a request of a namespace without rules.
func BenchmarkSecRuleNoRules(b *testing.B) {
	ctx := newTestContext(http.MethodGet, "/", "", nil)
	ctx.X().Namespace = "bench-none.example.com"
	plugin := &SecRule{store: secrules.New()}
	b.ReportAllocs()

	for b.Loop() {
		_ = plugin.Handle(ctx, benchNext())
	}
}

// BenchmarkSecRuleNoMatch rules evaluated, none matching.
func BenchmarkSecRuleNoMatch(b *testing.B) {
	benchSecRule(b, "bench-nomatch.example.com", neverMatches(), _block)
}

// BenchmarkSecRuleBlock a matching BLOCK rule: the event is logged and
// the 403 answered.
func BenchmarkSecRuleBlock(b *testing.B) {
	benchSecRule(b, "bench-block.example.com", &rulepb.Expression{
		OrCondition: []*rulepb.AndCondition{{}},
	}, _block)
}

// BenchmarkSecRuleActions a matching non-terminal rule: its action is
// applied and logged, then the chain goes on.
func BenchmarkSecRuleActions(b *testing.B) {
	benchSecRule(b, "bench-tag.example.com", &rulepb.Expression{
		OrCondition: []*rulepb.AndCondition{{}},
	}, _tag)
}

// BenchmarkSecRuleChallenge a matching CHALLENGE rule of a client
// without a pass: the challenge page is served.
func BenchmarkSecRuleChallenge(b *testing.B) {
	benchSecRule(b, "bench-challenge.example.com", &rulepb.Expression{
		OrCondition: []*rulepb.AndCondition{{}},
	}, _challenge)
}

func benchSecRule(
	b *testing.B, ns string, expr *rulepb.Expression, typ rulepb.ActionType,
) {
	b.Helper()

	store := secrules.New()
	err := store.Store(ns, []*secrulepb.SecRule{{
		Id:     "srule_bench",
		Status: typepb.Status_STATUS_ACTIVE,
		Expr:   expr,
		Action: &rulepb.Action{Type: typ},
	}})
	require.NoError(b, err)
	b.Cleanup(func() { store.Delete(ns) })

	ctx := newTestContext(http.MethodGet, "/", "", httptest.NewRecorder())
	ctx.X().Namespace = ns
	plugin := &SecRule{
		store: store,
		pow:   pow.New([]byte("bench-secret"), 1, time.Minute),
		log:   benchLog(typepb.LogType_LOG_TYPE_SECRULE),
	}
	b.ReportAllocs()

	for b.Loop() {
		_ = plugin.Handle(ctx, benchNext())
	}
}

// neverMatches is an expression no request of the benchmarks matches.
func neverMatches() *rulepb.Expression {
	return &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{
		Rules: []*rulepb.Rule{{Condition: &rulepb.Condition{
			Source:   rulepb.FieldSource_FIELD_SOURCE_PATH,
			Operator: rulepb.Operator_OPERATOR_EQ,
			Value:    structpb.NewStringValue("/never"),
		}}},
	}}}
}

// BenchmarkDomainBased the namespace and setting lookup of a request.
func BenchmarkDomainBased(b *testing.B) {
	const ns = "a.example.com"

	store := settings.New()
	store.Store(&edgepb.Setting{Ingress: &edgepb.Ingress{
		Resource: &edgepb.Resource{Domain: ns},
	}})
	b.Cleanup(func() { store.Delete(ns) })

	ctx := newTestContext(http.MethodGet, "/", "", nil)
	plugin := &DomainBased{settings: store}
	b.ReportAllocs()

	for b.Loop() {
		_ = plugin.Handle(ctx, benchNext())
	}
}

// BenchmarkMisdirected the SNI check of a request.
func BenchmarkMisdirected(b *testing.B) {
	ctx := newTestContext(http.MethodGet, "/", "", nil)
	b.ReportAllocs()

	for b.Loop() {
		_ = misdirected(ctx, "a.example.com")
	}
}

// BenchmarkSecRuleset a request through the OWASP CRS.
func BenchmarkSecRuleset(b *testing.B) {
	const ns = "bench-crs.example.com"

	store := secrulesets.New()
	err := store.Store(ns, &rulesetpb.Setting{
		Status: typepb.Status_STATUS_ACTIVE,
		Crs: &rulesetpb.Crs{
			Version: rulesetpb.CRSVersion_CRS_VERSION_V4_17_0,
			Mode:    rulesetpb.EngineMode_ENGINE_MODE_ON,
		},
	})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { store.Delete(ns) })

	plugin := &SecRuleset{
		store: store,
		log:   benchLog(typepb.LogType_LOG_TYPE_SECRULESET),
	}

	b.Run("clean", func(b *testing.B) {
		benchRuleset(b, plugin, ns, "/index.html?a=1")
	})
	b.Run("attack", func(b *testing.B) {
		benchRuleset(b, plugin, ns, "/?q=%3Cscript%3Ealert(1)%3C/script%3E")
	})
}

// BenchmarkSecRulesetNoWAF a request of a namespace without a WAF.
func BenchmarkSecRulesetNoWAF(b *testing.B) {
	plugin := &SecRuleset{store: secrulesets.New()}
	benchRuleset(b, plugin, "bench-nowaf.example.com", "/")
}

func benchRuleset(b *testing.B, p *SecRuleset, ns, target string) {
	b.Helper()
	b.ReportAllocs()

	for b.Loop() {
		ctx := newTestContext(
			http.MethodGet, target, "", httptest.NewRecorder())
		ctx.X().Namespace = ns
		_ = p.Handle(ctx, benchNext())
	}
}

// BenchmarkEvaluate a chain of matching and non-matching rules.
func BenchmarkEvaluate(b *testing.B) {
	ctx := newTestContext(http.MethodGet, "/", "", nil)
	chain := chainOf([]rule{
		{_none, false}, {_tag, true}, {_log, true},
		{_redirect, true}, {_challenge, true}, {_block, false},
	})
	var mr rulepb.MatchedRules
	var buf [_maxActions]*rulepb.Action
	b.ReportAllocs()

	for b.Loop() {
		_, _ = evaluate(ctx, chain, &mr, buf[:0])
	}
}

// BenchmarkApplyActions the non-terminal actions of a match.
func BenchmarkApplyActions(b *testing.B) {
	headers, _ := structpb.NewStruct(map[string]any{"X-Bench": "1"})
	tags, _ := structpb.NewStruct(map[string]any{"bench": "yes"})
	route, _ := structpb.NewStruct(map[string]any{_paramLocation: "/v2"})
	acts := []*rulepb.Action{
		{Type: _modifyHeader, Params: headers},
		{Type: _tag, Params: tags},
		{Type: _routeTo, Params: route},
	}
	ctx := newTestContext(http.MethodGet, "/", "", nil)
	b.ReportAllocs()

	for b.Loop() {
		applyActions(ctx, acts)
	}
}

// BenchmarkRedirect the REDIRECT answer of a rule.
func BenchmarkRedirect(b *testing.B) {
	params, _ := structpb.NewStruct(map[string]any{
		_paramLocation: "https://example.com/", _paramStatus: 301,
	})
	action := &rulepb.Action{Type: _redirect, Params: params}
	ctx := newTestContext(http.MethodGet, "/", "", httptest.NewRecorder())
	b.ReportAllocs()

	for b.Loop() {
		_, _ = redirect(ctx, action)
	}
}

// BenchmarkNewSecRuleEvent the event of a matched rule.
func BenchmarkNewSecRuleEvent(b *testing.B) {
	ctx := newTestContext(http.MethodGet, "/a?b=1", "", nil)
	mr := &rulepb.MatchedRules{Ids: []string{"srule_1"}, Names: []string{"n"}}
	b.ReportAllocs()

	for b.Loop() {
		e := newSecRuleEvent(ctx,
			rulepb.RuleService_RULE_SERVICE_SECRULE, _block, mr)
		_ = e.Close()
	}
}

// benchLog is a logger that drops what it is given.
func benchLog(kind typepb.LogType) szlog.LogCloser {
	return szlog.NewLogCloser("bench", kind, szlog.LevelFatal)
}

func benchNext() chains.Handler {
	return chains.HandlerFunc(func(szhttp.Context) error { return nil })
}
