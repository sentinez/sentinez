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
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szcrypto/pow"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/sentinez/sentinez/pkg/challenge"
	"github.com/sentinez/sentinez/pkg/pools/matchedrule"
	"github.com/sentinez/sentinez/pkg/pools/secrule"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/szlog"
)

// SecRuleId is the chain stage name of SecRule.
const SecRuleId = "RUL"

var (
	_ chains.Plugin                      = (*SecRule)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*SecRule)(nil)
)

type SecRule struct {
	log   szlog.LogCloser
	store *secrules.SecRule
	pow   *pow.Challenge
}

func (*SecRule) ID() string { return SecRuleId }

func (sr *SecRule) Provision(env *envplugin.Env) error {
	challenge, err := challenge.New(env.SecretKey())
	if err != nil {
		return fmt.Errorf("new challenge: %w", err)
	}

	sr.pow = challenge
	sr.store = env.Store().SecRule()
	sr.log = env.Logger(typepb.LogType_LOG_TYPE_SEC_RULE)
	return nil
}

// Handle runs the active rules of the namespace in priority order and
// applies the action matchAction picks: block ends the chain, challenge
// serves the proof-of-work page until the client holds a pass. A solved
// challenge posted to ChallengePath is answered here, before any rule.
func (sr *SecRule) Handle(ctx szhttp.Context, next chains.Handler) error {
	if challenge.IsVerify(ctx) {
		return sr.verify(ctx)
	}

	chain := sr.store.LoadContext(ctx)
	if len(chain) == 0 {
		return next.Handle(ctx)
	}

	matched := matchedrule.Acquire()
	defer func() { _ = matched.Close() }()

	action := matchAction(ctx, chain, matched.MatchedRules)
	if action == nil {
		return next.Handle(ctx)
	}

	sr.logMatched(ctx, action, matched.MatchedRules)

	switch action.GetType() {
	case rulepb.ActionType_ACTION_TYPE_BLOCK:
		return szhttp.Forbidden(ctx)
	case rulepb.ActionType_ACTION_TYPE_CHALLENGE:
		return sr.challenge(ctx, next)
	}

	return next.Handle(ctx)
}

// logMatched emits a secrule event for the request matched by mr.
func (sr *SecRule) logMatched(
	ctx szhttp.Context, action *rulepb.Action, mr *rulepb.MatchedRules,
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

// matchAction returns the action of the first matching rule with a
// terminal action (block), else of the first matching challenge, else of
// the last matching rule. A challenge does not stop the scan, so a later
// block still applies to clients holding a pass. A rule that does not
// match has no say: nil when no rule matches.
func matchAction(
	ctx szhttp.Context, chain []secrules.Entry, mr *rulepb.MatchedRules,
) *rulepb.Action {
	var action *rulepb.Action

	for _, e := range chain {
		if !e.Eval(ctx, mr) {
			continue
		}

		next := e.Rule.GetAction()
		switch {
		case next.GetType() == rulepb.ActionType_ACTION_TYPE_BLOCK:
			return next
		case action.GetType() != rulepb.ActionType_ACTION_TYPE_CHALLENGE:
			action = next
		}
	}

	return action
}

// challenge lets a client holding a valid pass through, and serves the
// challenge page to the others.
func (sr *SecRule) challenge(ctx szhttp.Context, next chains.Handler) error {
	if sr.pow.ValidPass(ctx) {
		return next.Handle(ctx)
	}

	ch := sr.pow.Make(ctx, sr.pow.CurrentSlot())

	return szhttp.Challenge(ctx, ch,
		sr.pow.Difficulty(), challenge.ChallengePath)
}

// verify checks the posted nonce, then sets the pass cookie and sends the
// client back to the page it was challenged on. A challenge that expired
// or was made for another client sends it back without a pass, so the
// page serves a fresh one.
func (sr *SecRule) verify(ctx szhttp.Context) error {
	form, err := url.ParseQuery(bytesconv.B2s(ctx.Body()))
	if err != nil {
		return szhttp.BadRequest(ctx)
	}

	nonce, err := strconv.ParseUint(form.Get("nonce"), 10, 64)
	if err != nil {
		return szhttp.BadRequest(ctx)
	}

	location := challenge.LocalPath(form.Get("redirect"))
	ctx.SetResponseHeader(szbytestr.HeaderLocation, []byte(location))

	if !sr.pow.Verify(ctx, form.Get("challenge"), nonce) {
		szlog.Debugf("secrule: challenge rejected: request %s ja4 %s",
			ctx.RequestId(), ctx.JA4())
		return ctx.String(http.StatusSeeOther, nil)
	}

	pass := sr.pow.IssuePass(ctx, challenge.PassTTL).String()
	ctx.SetResponseHeader(szbytestr.HeaderSetCookie, []byte(pass))

	return ctx.String(http.StatusSeeOther, nil)
}
