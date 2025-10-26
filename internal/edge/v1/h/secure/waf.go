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

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	rulecmn "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/common/v1"
	"github.com/sentinez/sentinez/core"
	"github.com/sentinez/sentinez/internal/edge/pkgs/chains"
	"github.com/sentinez/sentinez/internal/edge/pkgs/memory/wafengine"
	httpxhz "github.com/sentinez/sentinez/pkg/network/httpx/hz"
	"github.com/sentinez/sentinez/pkg/storage/cache/mem"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func NewWAF() *WAF {
	return &WAF{
		BaseHandler: chains.New(),
		logger: zlog.NewJSONLogger(
			edgepb.GetMetaEdgeServiceKey(),
			common.LogKind_LOG_KIND_WAF,
			zlog.LevelInfo,
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
func (w *WAF) Handle(ctx *httpxhz.Context) error {
	zlog.Debugf("[edge][%s] >>> visit WAF", ctx.GetReqID())

	waf := wafengine.GetEngine().LoadContext(ctx)
	if waf == nil {
		return w.HandleNext(ctx)
	}

	next := w.GetNext()
	if next == nil {
		return nil
	}

	rulesets := core.NewRulesets(ctx, waf)
	defer func() {
		rulesets.Final(func() { w.capture(ctx, rulesets) })
		rulesets.Release()
	}()

	if rulesets.IsRuleEngineOff() {
		return w.HandleNext(ctx)
	}

	// tx := httpsec.NewTransaction(waf, ctx)
	// defer httpsec.PostProcess(ctx, tx, w.capture)

	// if tx.IsRuleEngineOff() {
	// 	return w.HandleNext(ctx)
	// }

	// error for debuf WAF engine, not response
	// if err := httpsec.ProcessRequestHandler(ctx, tx); err != nil {
	// 	httpsec.DebugLogger(tx, err, "failed to process request")
	// 	return nil
	// }

	if err := rulesets.ExecIngress(ctx); err != nil {
		if ctx.StatusCode() == http.StatusForbidden {
			return httpxhz.Forbidden(ctx)
		}
	}

	err := w.HandleNext(ctx)

	if err := rulesets.ExecEgress(ctx); err != nil {
		if ctx.StatusCode() == http.StatusForbidden {
			return httpxhz.Forbidden(ctx)
		}
	}

	// error for debuf WAF engine, not response
	// if err := httpsec.ProcessResponseHandler(ctx, tx); err != nil {
	// 	httpsec.DebugLogger(tx, err, "failed to process response")
	// 	return nil
	// }

	return err
}

// nolint:funlen
func (w *WAF) capture(ctx *httpxhz.Context, rulesets *core.Rulesets) {

	interruption, matched, isInterrupted := rulesets.Matched()
	if !isInterrupted {
		return
	}

	if data, ok := w.cached.Get(httpxhz.GenContextKey(ctx)); ok {
		var event rulecmn.Event
		if err := event.UnmarshalVT(data); err != nil {
			return
		}

		event.RequestTime = ctx.Time().UnixMilli()
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

	event := &rulecmn.Event{
		RuleIds:       ruleIDs,
		Severities:    severities,
		Messages:      msgs,
		Path:          ctx.URI(),
		Score:         int32(score),
		Ip:            ctx.ClientIP(),
		RequestDomain: string(ctx.Host()),
		TransactionId: rulesets.GetTxId(),
		Service:       rulecmn.Service_SERVICE_RULE_CORE_RULESETS,
		Action:        rulecmn.Action_ACTION_DENY,
		RequestTime:   ctx.Time().UnixMilli(),
		HttpReqId:     ctx.GetReqID(),
		ContentType:   string(ctx.Unwrap().Request.Header.ContentType()),
	}

	w.logger.Info("rule engine ingress matched", event)
	data, _ := event.MarshalVT()
	w.cached.Set(httpxhz.GenContextKey(ctx), data)
}
