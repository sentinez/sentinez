// Copyright 2026 Duc-Hung Ho.
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
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sentinez/core/szcrypto/pow"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/sentinez/sentinez/pkg/challenge"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	_testDifficulty = 2

	_block     = rulepb.ActionType_ACTION_TYPE_BLOCK
	_challenge = rulepb.ActionType_ACTION_TYPE_CHALLENGE
	_log       = rulepb.ActionType_ACTION_TYPE_LOG
	_none      = rulepb.ActionType_ACTION_TYPE_UNSPECIFIED
)

// rule is one chain entry: its action and whether it matches.
type rule struct {
	action rulepb.ActionType
	match  bool
}

func TestMatchAction(t *testing.T) {
	tests := []struct {
		name      string
		giveRules []rule
		want      rulepb.ActionType
	}{
		{"no rules", nil, _none},
		{"no match", []rule{{_block, false}}, _none},
		{"last match", []rule{{_log, true}, {_log, true}}, _log},
		{"block wins", []rule{{_log, true}, {_block, true}, {_log, true}},
			_block},
		{"challenge kept", []rule{{_challenge, true}, {_log, true}},
			_challenge},
		{"block after challenge", []rule{{_challenge, true}, {_block, true}},
			_block},
		{"challenge after log", []rule{{_log, true}, {_challenge, true}},
			_challenge},
		{"unmatched block", []rule{{_challenge, true}, {_block, false}},
			_challenge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContext(http.MethodGet, "/", "", nil)
			mr := &rulepb.MatchedRules{}

			got := matchAction(ctx, chainOf(tt.giveRules), mr)

			assert.Equal(t, tt.want, got.GetType())
		})
	}
}

func TestSecRuleChallengeServesPage(t *testing.T) {
	sr := newTestSecRule()
	rec := httptest.NewRecorder()
	ctx := newTestContext(http.MethodGet, "/", "", rec)

	require.NoError(t, sr.challenge(ctx, failNext(t)))

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	wantAction := `action="` + challenge.ChallengePath + `"`
	assert.Contains(t, rec.Body.String(), wantAction)
}

func TestSecRuleChallengeValidPass(t *testing.T) {
	sr := newTestSecRule()
	ctx := newTestContext(http.MethodGet, "/", "", nil)
	pass := sr.pow.IssuePass(ctx, time.Minute)
	ctx.SetHeader([]byte("Cookie"), []byte(pass.Name+"="+pass.Value))

	called := false
	next := chains.HandlerFunc(func(szhttp.Context) error {
		called = true
		return nil
	})

	require.NoError(t, sr.challenge(ctx, next))
	assert.True(t, called)
}

func TestSecRuleVerifySolved(t *testing.T) {
	sr := newTestSecRule()
	ch := sr.pow.Make(
		newTestContext(http.MethodGet, "/", "", nil), sr.pow.CurrentSlot())
	body := url.Values{
		"challenge": {ch},
		"nonce":     {firstNonce(sr.pow, ch, true)},
		"redirect":  {"/shop?page=2"},
	}.Encode()

	rec := httptest.NewRecorder()
	ctx := newTestContext(http.MethodPost, challenge.ChallengePath, body, rec)

	require.True(t, challenge.IsVerify(ctx))
	require.NoError(t, sr.verify(ctx))

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/shop?page=2", rec.Header().Get("Location"))
	cookie := rec.Header().Get("Set-Cookie")
	assert.True(t, strings.HasPrefix(cookie, "pow-pass="))
	assert.Contains(t, cookie, "HttpOnly")
}

func TestSecRuleVerifyBadNonce(t *testing.T) {
	sr := newTestSecRule()
	body := url.Values{"challenge": {"ab"}, "nonce": {"x"}}.Encode()
	rec := httptest.NewRecorder()
	ctx := newTestContext(http.MethodPost, challenge.ChallengePath, body, rec)

	require.NoError(t, sr.verify(ctx))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestSecRuleVerifyRetry checks that a challenge that cannot be accepted
// sends the client back for a fresh one, without a pass.
func TestSecRuleVerifyRetry(t *testing.T) {
	sr := newTestSecRule()
	req := newTestContext(http.MethodGet, "/", "", nil)
	slot := sr.pow.CurrentSlot()
	current := sr.pow.Make(req, slot)
	expired := sr.pow.Make(req, slot-2)

	tests := []struct {
		name      string
		giveCh    string
		giveNonce string
	}{
		{"unsolved", current, firstNonce(sr.pow, current, false)},
		{"expired", expired, firstNonce(sr.pow, expired, true)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := url.Values{
				"challenge": {tt.giveCh},
				"nonce":     {tt.giveNonce},
				"redirect":  {"/login"},
			}.Encode()
			rec := httptest.NewRecorder()
			ctx := newTestContext(
				http.MethodPost, challenge.ChallengePath, body, rec)

			require.NoError(t, sr.verify(ctx))
			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Equal(t, "/login", rec.Header().Get("Location"))
			assert.Empty(t, rec.Header().Get("Set-Cookie"))
		})
	}
}

func chainOf(rules []rule) []secrules.Entry {
	chain := make([]secrules.Entry, 0, len(rules))
	for _, r := range rules {
		chain = append(chain, secrules.Entry{
			Eval: func(szhttp.Request, *rulepb.MatchedRules) bool {
				return r.match
			},
			Rule: &secrulepb.SecRule{Action: &rulepb.Action{Type: r.action}},
		})
	}

	return chain
}

func newTestSecRule() *SecRule {
	return &SecRule{
		pow: pow.New([]byte("test-secret"), _testDifficulty, time.Minute),
	}
}

// newTestContext builds a request from the same client each time, so a
// challenge made on one context verifies on another. A nil rec discards
// the response.
func newTestContext(
	method, target, body string, rec *httptest.ResponseRecorder,
) szhttp.Context {
	if rec == nil {
		rec = httptest.NewRecorder()
	}

	req := httptest.NewRequest(method, "https://a.example.com"+target,
		strings.NewReader(body))
	req.Header.Set("User-Agent", "Mozilla/5.0")

	return httpx.NewContext(req, rec)
}

func failNext(t *testing.T) chains.Handler {
	return chains.HandlerFunc(func(szhttp.Context) error {
		t.Error("next called")
		return nil
	})
}

// firstNonce returns the first nonce whose Check result on ch equals
// want.
func firstNonce(c *pow.Challenge, ch string, want bool) string {
	var nonce uint64
	for c.Check(ch, nonce) != want {
		nonce++
	}

	return strconv.FormatUint(nonce, 10)
}
