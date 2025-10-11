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
	"strconv"
	"time"

	"github.com/corazawaf/coraza/v3/types"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/net/waf/v1"
	"github.com/sentinez/sentinez/internal/common/chains"
	wafcache "github.com/sentinez/sentinez/internal/common/memory/waf"
	httpxhz "github.com/sentinez/sentinez/pkg/network/httpx/hz"
	httpxhzsec "github.com/sentinez/sentinez/pkg/network/httpx/hz/sec"
	"github.com/sentinez/sentinez/pkg/storage/cache/mem"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func NewWAF() *WAF {
	return &WAF{
		BaseHandler: chains.New(),
		logger: zlog.NewLoggingJSON(
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

func (w *WAF) Handle(ctx *httpxhz.Context) error {

	zlog.Info("[edge][handler] >>> WAF")

	waf := wafcache.GetWafCache().LoadContext(ctx)
	if waf == nil {
		return w.HandleNext(ctx)
	}

	next := w.GetNext()
	if next == nil {
		return nil
	}

	nx := httpxhzsec.WrapHandlerWithCallback(waf, next.Handle, w.callback)

	return nx(ctx)
}

// nolint:funlen
func (w *WAF) callback(ctx *httpxhz.Context, tx types.Transaction) {
	if !tx.IsInterrupted() {
		return
	}

	if data, ok := w.cached.Get(httpxhz.GenerateContextKey(ctx)); ok {
		var event waf.Event
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

	matched := tx.MatchedRules()
	for _, rule := range matched {
		if rule.Rule().ID() == tx.Interruption().RuleID {
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

	event := &waf.Event{
		RuleIds:       ruleIDs,
		Severities:    severities,
		Messages:      msgs,
		Path:          string(ctx.Request.URI().RequestURI()),
		Score:         int32(score),
		Ip:            ctx.ClientIP(),
		RequestDomain: string(ctx.Host()),
		TransactionId: tx.ID(),
		Service:       waf.Service_SERVICE_WAF_RULESETS,
		Action:        waf.Action_ACTION_DENY,
		RequestTime:   ctx.Time().UnixMilli(),
		HttpReqId:     httpxhz.GetContextIdentify(ctx),
		ContentType:   string(ctx.Request.Header.ContentType()),
	}

	w.logger.Info("rule engine ingress matched", event)
	data, _ := event.MarshalVT()
	w.cached.Set(httpxhz.GenerateContextKey(ctx), data)
}
