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
	"github.com/sentinez/pkg/bytesconv"
	"github.com/sentinez/pkg/szlog"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/memory/rulesets"
	"github.com/sentinez/sentinez/pkg/pools/coreruleset"
)

// WAFId is the chain stage name of WAF.
const WAFId = "WAF"

var (
	_ chains.Plugin                      = (*Rulesets)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*Rulesets)(nil)
)

type Rulesets struct {
	log   szlog.LogCloser
	store *rulesets.Rulesets
}

func (*Rulesets) ID() string { return WAFId }

func (rs *Rulesets) Provision(env *envplugin.Env) error {
	rs.store = env.Store().Rulesets()
	rs.log = env.Logger(typepb.LogType_LOG_TYPE_CORE_RULESETS)
	return nil
}

// nolint:funlen
func (rs *Rulesets) Handle(ctx szhttp.Context, next chains.Handler) error {
	// szlog.Debug("[edge] >>> visit WAF")

	waf, ok := rs.store.LoadContext(ctx)
	if !ok {
		return next.Handle(ctx)
	}

	ruleset := szcrs.NewRulesets(ctx, waf)
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
func (rs *Rulesets) capture(ctx szhttp.Context, ruleset *szcrs.Rulesets) {
	matched, it, blocked := ruleset.Matched()
	if it == nil {
		return
	}

	event := coreruleset.Acquire()
	fillDetections(event, matched, it.RuleID)

	event.Behavior = rulepb.RuleBehavior_RULE_BEHAVIOR_LOG
	if blocked {
		event.Behavior = rulepb.RuleBehavior_RULE_BEHAVIOR_DENY
	}

	inbound, outbound := ruleset.AnomalyScore()
	event.TotalScore = int32(inbound + outbound)
	event.Service = rulepb.RuleService_RULE_SERVICE_CORE_RULESETS
	event.MatchedTime = ctx.RequestTime().UnixMilli()

	event.RequestId = ctx.RequestId()
	event.RequestDomain = bytesconv.B2s(ctx.Host())
	event.RequestPath = bytesconv.B2s(ctx.URI())
	event.RequestIp = bytesconv.B2s(ctx.RequestIP())
	event.RequestQuery = bytesconv.B2s(ctx.QueryStr())
	event.RequestScheme = ctx.Scheme()
	event.ResourceId = ctx.ResourceId()
	event.TenantId = ctx.TenantId()

	rs.log.Info("rulesets: matched", event, event)
}

// fillDetections sets the detecting rules of event: the matched rules with
// a message, but the framework rules and the interrupting rule.
func fillDetections(
	event *coreruleset.CoreRuleEvent,
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
