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

package edge

import (
	"context"
	"strconv"
	"time"

	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/net/waf/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/logic"
	"github.com/sentinez/sentinez/internal/edge/v1/routing"
	httpxf1mdw "github.com/sentinez/sentinez/pkg/core/net/httpx/f1/middleware"
	"github.com/valyala/fasthttp"
)

func (s *Server) bootloader(_ context.Context) error {

	protected := httpxf1mdw.ProtectedWithCallback(
		s.flag.GetRulePath(), s.rulesCallback)

	host := logic.Host(s.flag.GetHost())

	s.core.Use(host)
	s.core.Use(protected)

	return routing.Serve(s.config, s.core)
}

// nolint:funlen
func (s *Server) rulesCallback(ctx *fasthttp.RequestCtx, tx types.Transaction) {
	if !tx.IsInterrupted() {
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

	s.logger.Info("rule engine ingress matched rule", &waf.Event{
		RuleIds:       ruleIDs,
		Severities:    severities,
		Messages:      msgs,
		RequestPath:   string(ctx.RequestURI()),
		Score:         int32(score),
		RequestIp:     ctx.RemoteIP().String(),
		RequestDomain: string(ctx.Host()),
		TransactionId: tx.ID(),
		Service:       waf.Service_SERVICE_RULESETS,
		Action:        waf.Action_ACTION_DENY,
		RequestTime:   time.Now().UTC().UnixMilli(),
	})
}
