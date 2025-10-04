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
	httpxhz "github.com/sentinez/sentinez/pkg/core/net/httpx/hz"
	httpxhzmdw "github.com/sentinez/sentinez/pkg/core/net/httpx/hz/mdw"
	"github.com/sentinez/sentinez/pkg/infra/cache/mem"
	"strconv"
	"time"

	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/net/waf/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	logger zlog.Logger
	cached *mem.Cache[[]byte]
)

func WAFHandler(rulePath string,
) func(httpxhz.RequestHandler) httpxhz.RequestHandler {
	logger = zlog.NewLoggingJSON(
		edge.GetMetaEdgeServiceKey(),
		common.LogKind_LOG_KIND_WAF,
		zlog.LevelInfo,
	)

	cached = mem.New[[]byte](time.Second*30, time.Second*31)

	protected := httpxhzmdw.ProtectedWithCallback(rulePath, rulesCallback)
	return protected
}

// nolint:funlen
func rulesCallback(ctx *httpxhz.Context, tx types.Transaction) {
	if !tx.IsInterrupted() {
		return
	}

	if data, ok := cached.Get(httpxhz.GenerateContextKey(ctx)); ok {
		var event waf.Event
		if err := event.UnmarshalVT(data); err != nil {
			return
		}

		event.RequestTime = ctx.Time().UnixMilli()
		logger.Info("cache hit: rule engine ingress matched", &event)
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

	logger.Info("rule engine ingress matched", event)
	data, _ := event.MarshalVT()
	cached.Set(httpxhz.GenerateContextKey(ctx), data)
}
