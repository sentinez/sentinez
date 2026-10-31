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
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/envplugin"
	memlimiter "github.com/sentinez/sentinez/internal/memory/ratelimiter"
	"github.com/sentinez/sentinez/pkg/pools/matchedrule"
	"github.com/sentinez/shared/szlog"
)

// LimiterId is the chain stage name of Limiter.
const LimiterId = "LMT"

var (
	_ chains.Plugin                      = (*Limiter)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*Limiter)(nil)
)

type Limiter struct {
	store *memlimiter.Limiter
	log   szlog.LogCloser
}

func (*Limiter) ID() string { return LimiterId }

func (l *Limiter) Provision(env *envplugin.Env) error {
	l.store = env.Store().Limiter()
	l.log = env.Logger(typepb.LogType_LOG_TYPE_RATE_LIMIT)
	return nil
}

// Handle runs the limiters of the namespace in priority order, like the
// secrules: each one whose expression matches counts the request, keyed by
// client IP, and the first one over its limit rejects it.
func (l *Limiter) Handle(ctx szhttp.Context, next chains.Handler) error {
	chain := l.store.LoadContext(ctx)
	if len(chain) == 0 {
		return next.Handle(ctx)
	}

	matched := matchedrule.Acquire()
	defer func() { _ = matched.Close() }()

	key := string(ctx.RequestIP())
	for _, e := range chain {
		if e.Eval(ctx, matched.MatchedRules) && !e.Limiter.Allow(key) {
			l.logExceeded(ctx, e.Rule.GetAction(), matched.MatchedRules)
			return szhttp.TooManyRequests(ctx)
		}
	}

	return next.Handle(ctx)
}

// logExceeded emits a rate limit event for the request rejected with
// action, listing the limiters matched by mr, the rejecting one last.
func (l *Limiter) logExceeded(
	ctx szhttp.Context, action *rulepb.Action, mr *rulepb.MatchedRules,
) {
	event := newRuleEvent(ctx,
		rulepb.RuleService_RULE_SERVICE_RATE_LIMIT, action, mr)

	// Info releases the event back to the pool.
	l.log.Info("ratelimiter: exceeded", event, event)
}
