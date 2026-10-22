// Copyright 2025 Sentinéz Labs.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
	memlimiter "github.com/sentinez/sentinez/internal/memory/ratelimiter"
	"github.com/sentinez/sentinez/pkg/pools/matchedrule"
	"github.com/sentinez/sentinez/pkg/pools/secrule"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/zlog"
)

func NewLimiter(
	store *memory.MemStore, opts ...option.Option) corechains.ChainNode {

	o := option.New(opts...)

	return &Limiter{
		Node:  corechains.NewNode(),
		store: store,
		log: zlog.NewLogCloser(
			o.Named,
			typepb.LogType_LOG_TYPE_RATE_LIMIT,
			o.Level,
			o.OTLP(),
		),
	}
}

type Limiter struct {
	*corechains.Node
	store *memory.MemStore
	log   zlog.LogCloser
}

// Handle counts the request against every limiter whose expression matches
// it, keyed by client IP. The request is rejected by the first limiter
// over its limit.
func (l *Limiter) Handle(ctx corehttp.Context) error {
	key := string(ctx.RequestIP())

	for _, e := range l.store.Limiter().LoadContext(ctx) {
		if e.Eval(ctx, nil) && !e.Limiter.Allow(key) {
			l.logExceeded(ctx, e)
			return corehttp.TooManyRequests(ctx)
		}
	}

	return l.HandleNext(ctx)
}

// logExceeded emits a rate limit event for the request rejected by e: the
// limiter rule first, then the conditions of its expression that matched.
func (l *Limiter) logExceeded(ctx corehttp.Context, e memlimiter.Entry) {
	// Eval again to collect the matched conditions, only on rejection so
	// that the common path stays allocation free.
	mr := matchedrule.Acquire()
	defer func() { _ = mr.Close() }()

	e.Eval(ctx, mr.MatchedRules)

	event := secrule.Acquire()

	event.Action = e.Rule.GetAction()
	event.MatchedSecruleIds = append(event.MatchedSecruleIds, e.Rule.GetId())
	event.MatchedSecruleIds = append(event.MatchedSecruleIds, mr.GetIds()...)
	event.Messages = append(event.Messages, e.Rule.GetName())
	event.Messages = append(event.Messages, mr.GetNames()...)
	event.Service = rulepb.RuleService_RULE_SERVICE_RATE_LIMIT
	event.MatchedTime = ctx.RequestTime().UnixMilli()

	event.RequestId = ctx.RequestId()
	event.RequestDomain = bytesconv.B2s(ctx.Host())
	event.RequestPath = bytesconv.B2s(ctx.URI())
	event.RequestIp = bytesconv.B2s(ctx.RequestIP())
	event.RequestQuery = bytesconv.B2s(ctx.QueryStr())

	// Info releases the event back to the pool.
	l.log.Info("ratelimiter: exceeded", event, event)
}
