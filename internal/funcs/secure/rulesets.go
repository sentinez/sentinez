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

	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	corers "github.com/sentinez/core/rulesets"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/sentinez/pkg/pools/coreruleset"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/zlog"
)

func NewWAF(
	store *memory.MemStore, opts ...option.Option) corechains.ChainNode {

	o := option.New(opts...)

	return &WAF{
		Node: corechains.NewNode(),
		log: zlog.NewLogCloser(
			o.Named,
			typepb.LogType_LOG_TYPE_WAF,
			o.Level,
			o.OTLP(),
		),
		store: store,
	}
}

type WAF struct {
	*corechains.Node
	log   zlog.LogCloser
	store *memory.MemStore
}

// nolint:funlen
func (w *WAF) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit WAF")

	waf, ok := w.store.Rulesets().LoadContext(ctx)
	if !ok {
		return w.HandleNext(ctx)
	}

	next := w.GetNext()
	if next == nil {
		return nil
	}

	ruleset := corers.NewRulesets(ctx, waf)
	defer func() {
		ruleset.Final(func() { w.capture(ctx, ruleset) })
		ruleset.Release()
	}()

	if ruleset.IsRuleEngineOff() {
		return w.HandleNext(ctx)
	}

	if err := ruleset.ExecIngress(ctx); err != nil {
		if ctx.StatusCode() == http.StatusForbidden {
			return corehttp.Forbidden(ctx)
		}
	}

	err := w.HandleNext(ctx)

	if err := ruleset.ExecEgress(ctx); err != nil {
		if ctx.StatusCode() == http.StatusForbidden {
			return corehttp.Forbidden(ctx)
		}
	}

	return err
}

// nolint:funlen
func (w *WAF) capture(ctx corehttp.Context, ruleset *corers.Rulesets) {

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

	w.log.Info("rulesets: matched", event, event)
}
