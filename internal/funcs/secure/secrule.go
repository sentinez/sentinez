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
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/sentinez/sentinez/pkg/pools/matchedrule"
	"github.com/sentinez/sentinez/pkg/pools/secrule"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/zlog"
)

func NewSecRule(
	store *memory.MemStore, opts ...option.Option) corechains.ChainNode {

	o := option.New(opts...)

	return &SecRule{
		Node:  corechains.NewNode(),
		store: store,
		log: zlog.NewLogCloser(
			o.Named,
			typepb.LogType_LOG_TYPE_SEC_RULE,
			o.Level,
			o.OTLP(),
		),
	}
}

type SecRule struct {
	*corechains.Node
	log   zlog.LogCloser
	store *memory.MemStore
}

// Handle runs the active rules of the namespace in priority order. The
// first matching rule with a terminal action (block) ends the chain; other
// matches fall through to the next rule.
func (sr *SecRule) Handle(ctx corehttp.Context) error {
	chain := sr.store.SecRule().LoadContext(ctx)
	if len(chain) == 0 {
		return sr.HandleNext(ctx)
	}

	matched := matchedrule.Acquire()
	defer func() { _ = matched.Close() }()

	action := matchAction(ctx, chain, matched.MatchedRules)
	if action == nil {
		return sr.HandleNext(ctx)
	}

	sr.logMatched(ctx, action, matched.MatchedRules)

	if action.GetType() == rulepb.ActionType_ACTION_TYPE_BLOCK {
		return corehttp.Forbidden(ctx)
	}

	return sr.HandleNext(ctx)
}

// logMatched emits a secrule event for the request matched by mr.
func (sr *SecRule) logMatched(
	ctx corehttp.Context, action *rulepb.Action, mr *rulepb.MatchedRules,
) {
	event := secrule.Acquire()

	event.Action = action
	event.MatchedSecruleIds = append(event.MatchedSecruleIds, mr.GetIds()...)
	event.Messages = append(event.Messages, mr.GetNames()...)
	event.Service = rulepb.RuleService_RULE_SERVICE_SEC_RULE
	event.MatchedTime = ctx.RequestTime().UnixMilli()

	event.RequestId = ctx.RequestId()
	event.RequestDomain = bytesconv.B2s(ctx.Host())
	event.RequestPath = bytesconv.B2s(ctx.URI())
	event.RequestIp = bytesconv.B2s(ctx.RequestIP())
	event.RequestQuery = bytesconv.B2s(ctx.QueryStr())

	// Info releases the event back to the pool.
	sr.log.Info("secrule: matched", event, event)
}

// matchAction the action of the first matching rule with a terminal
// action (block), else of the last matching rule. A rule that does not
// match has no say: nil when no rule matches.
func matchAction(
	ctx corehttp.Context, chain []secrules.Entry, mr *rulepb.MatchedRules,
) *rulepb.Action {
	var action *rulepb.Action

	for _, e := range chain {
		if !e.Eval(ctx, mr) {
			continue
		}

		action = e.Rule.GetAction()
		if action.GetType() == rulepb.ActionType_ACTION_TYPE_BLOCK {
			break
		}
	}

	return action
}
