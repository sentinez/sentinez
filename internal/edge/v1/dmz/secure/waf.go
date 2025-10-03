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

	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/net/waf/v1"
	httpxf1 "github.com/sentinez/sentinez/pkg/core/net/httpx/f1"
	httpxf1mdw "github.com/sentinez/sentinez/pkg/core/net/httpx/f1/middleware"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	logger zlog.Logger
	// cached *mem.Cache[[]byte]
)

func WAFHandler(rulePath string,
) func(httpxf1.RequestHandler) httpxf1.RequestHandler {
	logger = zlog.NewLoggingJSON(
		edge.GetMetaEdgeServiceKey(),
		common.LogKind_LOG_KIND_WAF,
		zlog.LevelInfo,
	)

	// cached = mem.New[[]byte](time.Second*30, time.Second*31)

	protected := httpxf1mdw.ProtectedWithCallback(rulePath, rulesCallback)
	return protected
}

// nolint:funlen
func rulesCallback(ctx *httpxf1.Context, tx types.Transaction) {
	if !tx.IsInterrupted() {
		return
	}

	// if data, ok := cached.Get(httpxf1.GenerateContextKey(ctx)); ok {
	// 	var event waf.Event
	// 	if err := event.UnmarshalVT(data); err != nil {
	// 		return
	// 	}

	// 	event.RequestTime = ctx.Time().UnixMilli()
	// 	logger.Info("cache hit: rule engine ingress matched", &event)
	// 	return
	// }

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
		Path:          string(ctx.RequestURI()),
		Score:         int32(score),
		Ip:            ctx.RemoteIP().String(),
		RequestDomain: string(ctx.Host()),
		TransactionId: tx.ID(),
		Service:       waf.Service_SERVICE_WAF_RULESETS,
		Action:        waf.Action_ACTION_DENY,
		RequestTime:   ctx.Time().UnixMilli(),
		HttpReqId:     httpxf1.GetContextIdentify(ctx),
		ContentType:   string(ctx.Request.Header.ContentType()),
	}

	logger.Info("rule engine ingress matched", event)
	// data, _ := event.MarshalVT()
	// cached.Set(httpxf1.GenerateContextKey(ctx), data)
}
