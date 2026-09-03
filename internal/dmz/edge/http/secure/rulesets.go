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

	"github.com/sentinez/core/common/bytestr"
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	corers "github.com/sentinez/core/rulesets"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/sentinez/pkg/pools/ruleevent"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/zlog"
)

func NewWAF(logLevel zlog.Level, store *memory.MemStore) corechains.ChainNode {
	return &WAF{
		Node: corechains.NewNode(),
		log: zlog.NewLogCloser(edgepb.GetMetaEdgeServiceKey(),
			typepb.LogKind_LOG_KIND_WAF, logLevel,
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

	event := ruleevent.Acquire()

	var (
		ruleIDs    []int32
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
			ruleIDs = append(ruleIDs, int32(rule.Rule().ID()))
			severities = append(severities, rule.Rule().Severity().String())
			msgs = append(msgs, rule.Message())
		}
	}

	event.RuleIds = ruleIDs
	event.Severities = severities
	event.Messages = msgs
	event.Path = bytesconv.B2s(ctx.URI())
	event.Score = int32(score)
	event.Ip = bytesconv.B2s(ctx.RequestIP())
	event.RequestDomain = bytesconv.B2s(ctx.Host())
	event.TransactionId = ruleset.GetTxId()
	event.Service = rulepb.RuleService_RULE_SERVICE_CORE_RULESETS
	event.Behavior = rulepb.RuleBehavior_RULE_BEHAVIOR_DENY
	event.RequestTime = ctx.RequestTime().UnixMilli()
	event.HttpReqId = ctx.RequestId()
	event.ContentType = bytesconv.B2s(ctx.Header(bytestr.HeaderContentType))

	w.log.Info("rulesets: matched", event, event)
}
