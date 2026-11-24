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
	"github.com/sentinez/contrib/pkg/bytesconv"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/pkg/szlog"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/memory/secratelimits"
	"github.com/sentinez/sentinez/pkg/pools/matchedrule"
)

// SecLimiterId is the chain stage name of SecLimiter.
const SecLimiterId = "secratelimit"

var (
	_ chains.Plugin                      = (*SecLimiter)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*SecLimiter)(nil)
)

// _rateLimited the action of a rate limit event, whatever the limiter is
// configured with: the request is always answered 429. Read only, it is
// shared by the events (Release drops it without touching it).
var _rateLimited = &rulepb.Action{
	Type: rulepb.ActionType_ACTION_TYPE_RATE_LIMIT,
}

type SecLimiter struct {
	store *secratelimits.Limiter
	log   szlog.LogCloser
}

func (*SecLimiter) ID() string { return SecLimiterId }

func (l *SecLimiter) Provision(env *envplugin.Env) error {
	l.store = env.Store().Limiter()
	l.log = env.Logger(typepb.LogType_LOG_TYPE_SECRATELIMIT)
	return nil
}

// Handle runs the limiters of the namespace in priority order, like the
// secrules: each one whose expression matches counts the request, keyed by
// client IP, and the first one over its limit rejects it.
func (l *SecLimiter) Handle(ctx szhttp.Context, next chains.Handler) error {
	chain := l.store.LoadContext(ctx)
	if len(chain) == 0 {
		return next.Handle(ctx)
	}

	matched := matchedrule.Acquire()
	defer func() { _ = matched.Close() }()

	run := scriptRun{ctx: ctx}
	defer run.Close()

	// not a copy: Allow copies the key it keeps
	ip := bytesconv.B2s(ctx.RequestIP())
	for _, e := range chain {
		key, ok := counted(&run, e, ip, matched.MatchedRules)
		if ok && !e.Limiter.Allow(key) {
			l.logExceeded(ctx, matched.MatchedRules)
			return szhttp.TooManyRequests(ctx)
		}
	}

	return next.Handle(ctx)
}

// counted the key the request is counted under by e, ok false when it is
// not: the client IP for an expr, the one its script returns otherwise.
func counted(run *scriptRun, e secratelimits.Entry, ip string,
	mr *rulepb.MatchedRules) (key string, ok bool) {

	if e.Script == nil {
		return ip, e.Eval(run.ctx, mr)
	}

	key, ok = run.key(e.Script, ip)
	if ok {
		matched(mr, e.Rule)
	}

	return key, ok
}

// logExceeded emits a rate limit event for the rejected request, listing
// the limiters matched by mr, the rejecting one last.
func (l *SecLimiter) logExceeded(ctx szhttp.Context, mr *rulepb.MatchedRules) {
	event := newSecRuleEvent(ctx,
		rulepb.RuleService_RULE_SERVICE_SECRATELIMIT, _rateLimited, mr)

	// Info releases the event back to the pool.
	l.log.Info("ratelimiter: exceeded", event, event)
}
