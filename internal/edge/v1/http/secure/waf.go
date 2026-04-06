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
	"time"

	corehttp "github.com/sentinez/core/http"
	corers "github.com/sentinez/core/rulesets"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	ruleeventpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleevent/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/shared/chains"
	"github.com/sentinez/sentinez/internal/shared/mem/wafengine"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
	"github.com/sentinez/sentinez/pkg/storage/cache/mem"
	"github.com/sentinez/shared/zlog"
)

func NewWAF(logLevel zlog.Level) chains.Handler {
	return &WAF{
		BaseHandler: chains.New(),
		logger: zlog.NewJSONLogger(edgepb.GetMetaEdgeServiceKey(),
			typepb.LogKind_LOG_KIND_WAF, logLevel,
		),
		cached: mem.New[[]byte](time.Second*30, time.Second*31),
	}
}

type WAF struct {
	*chains.BaseHandler
	logger zlog.Logger
	cached *mem.Cache[[]byte]
}

// nolint:funlen
func (w *WAF) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit WAF")

	waf := wafengine.GetEngine().LoadContext(ctx)
	if waf == nil {
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
			return httpxcmn.Forbidden(ctx)
		}
	}

	err := w.HandleNext(ctx)

	if err := ruleset.ExecEgress(ctx); err != nil {
		if ctx.StatusCode() == http.StatusForbidden {
			return httpxcmn.Forbidden(ctx)
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

	if data, ok := w.cached.Get(httpxcmn.GenContextKey(ctx)); ok {
		var event ruleeventpb.Event
		if err := event.UnmarshalVT(data); err != nil {
			return
		}

		event.RequestTime = ctx.RequestTime().UnixMilli()
		w.logger.Info("cache hit: rule engine ingress matched", &event)
		return
	}

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

	event := &ruleeventpb.Event{
		RuleIds:       ruleIDs,
		Severities:    severities,
		Messages:      msgs,
		Path:          ctx.URI(),
		Score:         int32(score),
		Ip:            ctx.RequestIP(),
		RequestDomain: ctx.Host(),
		TransactionId: ruleset.GetTxId(),
		Service:       ruleeventpb.Service_SERVICE_RULE_CORE_RULESETS,
		Action:        ruleeventpb.Action_ACTION_DENY,
		RequestTime:   ctx.RequestTime().UnixMilli(),
		HttpReqId:     ctx.RequestId(),
		ContentType:   ctx.Header(corehttp.HeaderContentType),
	}

	w.logger.Info("[rulesets] [matched]", event)
	data, _ := event.MarshalVT()
	w.cached.Set(httpxcmn.GenContextKey(ctx), data)
}
