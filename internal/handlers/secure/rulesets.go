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
	"strconv"

	"github.com/sentinez/core/szhttp/chains"

	"github.com/sentinez/core/szcrs"
	"github.com/sentinez/core/szhttp"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/memory/rulesets"
	"github.com/sentinez/sentinez/pkg/pools/coreruleset"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/szlog"
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
	rs.log = env.Logger(typepb.LogType_LOG_TYPE_WAF)
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

// nolint:funlen
func (rs *Rulesets) capture(ctx szhttp.Context, ruleset *szcrs.Rulesets) {

	interruption, matched, isInterrupted := ruleset.Matched()
	if !isInterrupted {
		return
	}

	var (
		ruleIds    []int32
		severities []string
		msgs       []string
		score      int
	)

	for _, rule := range matched {
		if rule.Rule().ID() == interruption.RuleID {
			score, _ = strconv.Atoi(rule.Data())
			continue
		}

		rule.Rule().Accuracy()

		if rule.Message() != "" {
			ruleIds = append(ruleIds, int32(rule.Rule().ID()))
			severities = append(severities, rule.Rule().Severity().String())
			msgs = append(msgs, rule.Message())
		}
	}

	event := coreruleset.Acquire()

	event.Id = ruleset.GetTxId()
	event.MatchedRuleIds = ruleIds
	event.Severities = severities
	event.Messages = msgs
	event.TotalScore = int32(score)
	event.Service = rulepb.RuleService_RULE_SERVICE_WAF_RULESETS
	event.Behavior = rulepb.RuleBehavior_RULE_BEHAVIOR_DENY
	event.MatchedTime = ctx.RequestTime().UnixMilli()

	event.RequestId = ctx.RequestId()
	event.RequestDomain = bytesconv.B2s(ctx.Host())
	event.RequestPath = bytesconv.B2s(ctx.URI())
	event.RequestIp = bytesconv.B2s(ctx.RequestIP())
	event.RequestQuery = bytesconv.B2s(ctx.QueryStr())

	rs.log.Info("rulesets: matched", event, event)
}
