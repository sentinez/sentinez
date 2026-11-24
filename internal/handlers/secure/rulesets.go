// Copyright 2025 Duc-Hung Ho.
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

	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/core/szcrs"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/core/szruleset"
	"github.com/sentinez/pkg/szlog"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/memory/secrulesets"
	secrulesetpool "github.com/sentinez/sentinez/pkg/pools/secruleset"
)

// SecRulesetId is the chain stage name of SecRuleset.
const SecRulesetId = "secruleset"

var (
	_ chains.Plugin                      = (*SecRuleset)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*SecRuleset)(nil)
)

type SecRuleset struct {
	log   szlog.LogCloser
	store *secrulesets.Rulesets
}

func (*SecRuleset) ID() string { return SecRulesetId }

func (rs *SecRuleset) Provision(env *envplugin.Env) error {
	rs.store = env.Store().Rulesets()
	rs.log = env.Logger(typepb.LogType_LOG_TYPE_SECRULESET)
	return nil
}

// nolint:funlen
func (rs *SecRuleset) Handle(ctx szhttp.Context, next chains.Handler) error {
	// szlog.Debug("[edge] >>> visit WAF")

	waf, ok := rs.store.LoadContext(ctx)
	if !ok {
		return next.Handle(ctx)
	}

	ruleset := szruleset.NewRulesets(ctx, waf)
	defer func() {
		ruleset.Final(func() { rs.capture(ctx, ruleset) })
		ruleset.Release()
	}()

	if ruleset.IsRuleEngineOff() {
		return next.Handle(ctx)
	}

	if err := ruleset.ExecIngress(ctx); err != nil {
		if ctx.StatusCode() == http.StatusForbidden {
			return szhttp.Forbidden(ctx)
		}
	}

	err := next.Handle(ctx)

	if err := ruleset.ExecEgress(ctx); err != nil {
		if ctx.StatusCode() == http.StatusForbidden {
			return szhttp.Forbidden(ctx)
		}
	}

	return err
}

// capture logs the rules of an interrupted request: DENY when the engine
// blocked it, LOG when the engine is DetectionOnly and would have blocked
// it. Only DetectionOnly reports an interruption it did not perform.
func (rs *SecRuleset) capture(ctx szhttp.Context, ruleset *szruleset.Rulesets) {
	matched, it, blocked := ruleset.Matched()
	if it == nil {
		return
	}

	event := secrulesetpool.Acquire()
	fillDetections(event, matched, it.RuleID)

	event.Behavior = rulepb.RuleBehavior_RULE_BEHAVIOR_LOG
	if blocked {
		event.Behavior = rulepb.RuleBehavior_RULE_BEHAVIOR_DENY
	}

	inbound, outbound := szcrs.AnomalyScore(ruleset)
	event.TotalScore = int32(inbound + outbound)
	event.Service = rulepb.RuleService_RULE_SERVICE_SECRULESET
	event.MatchedTime = ctx.RequestTime().UnixMilli()

	// exported after the request: Keep copies what does not outlive it
	event.RequestId = ctx.RequestId()
	event.RequestDomain = szhttp.Keep(ctx, ctx.Host())
	event.RequestPath = szhttp.Keep(ctx, ctx.URI())
	event.RequestIp = szhttp.Keep(ctx, ctx.RequestIP())
	event.RequestQuery = szhttp.Keep(ctx, ctx.QueryStr())
	event.RequestScheme = ctx.Scheme()
	event.ResourceId = ctx.ResourceId()
	event.TenantId = ctx.TenantId()

	rs.log.Info("rulesets: matched", event, event)
}

// fillDetections sets the detecting rules of event: the matched rules with
// a message, but the framework rules and the interrupting rule.
func fillDetections(
	event *secrulesetpool.SecRulesetEvent,
	matched []types.MatchedRule,
	interruptId int,
) {
	for _, rule := range matched {
		id := rule.Rule().ID()
		if id == interruptId || szcrs.IsFrameworkRule(id) ||
			rule.Message() == "" {
			continue
		}

		event.MatchedRuleIds = append(event.MatchedRuleIds, int32(id))
		event.Severities = append(
			event.Severities, rule.Rule().Severity().String(),
		)
		event.Messages = append(event.Messages, rule.Message())
	}
}
