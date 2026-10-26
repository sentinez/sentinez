// Copyright 2026 Duc-Hung Ho.
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

package pow

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sentinez/core/szcontext"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	"github.com/sentinez/core/szhttp/httpreq"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	netpb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	_testSlot = int64(29_000_000)
	_testJA4  = "t13d1516h2_8daaf6152771_02713d6af862"
)

var (
	_hex64   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	_passFmt = regexp.MustCompile(`^(\d+)\.([0-9a-f]{64})$`)
)

// fixture is one challenge plus the client attributes it binds to.
type fixture struct {
	challenge *Challenge
	ip        string
	ua        string
	ja4       string
	connId    string
	cookie    string
}

func newFixture() *fixture {
	return &fixture{
		challenge: newChallenge(0),
		ip:        "203.0.113.7",
		ua:        "Mozilla/5.0",
		ja4:       _testJA4,
		connId:    "abc123",
	}
}

func (f *fixture) request() szhttp.Request {
	ctx := szcontext.WithTransportValue(
		context.Background(), &netpb.Transport{ConnId: f.connId})

	return httpreq.NewRequestContext(ctx, &httppb.Request{
		ClientIp:    f.ip,
		Fingerprint: f.ja4,
		Headers: []*httppb.RequestHeader{
			header(httpconst.HeaderUserAgent, f.ua),
			header(httpconst.HeaderCookie, f.cookie),
		},
	})
}

func TestChallengeCurrentSlot(t *testing.T) {
	c := newChallenge(0)

	want := time.Now().Unix() / 60
	assert.InDelta(t, want, c.CurrentSlot(), 1)
}

func TestChallengeMakeDeterministic(t *testing.T) {
	f := newFixture()

	got := f.challenge.Make(f.request(), _testSlot)

	assert.Regexp(t, _hex64, got)
	assert.Equal(t, got, f.challenge.Make(f.request(), _testSlot))
}

func TestChallengeMakeBinding(t *testing.T) {
	f := newFixture()
	base := f.challenge.Make(f.request(), _testSlot)

	for _, tt := range bindingTests() {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			tt.giveMutate(f)

			got := f.challenge.Make(f.request(), _testSlot)

			assert.Equal(t, tt.wantSame, got == base)
		})
	}
}

func TestChallengeMakeSlot(t *testing.T) {
	f := newFixture()

	assert.NotEqual(t,
		f.challenge.Make(f.request(), _testSlot),
		f.challenge.Make(f.request(), _testSlot+1))
}

func TestChallengeCheck(t *testing.T) {
	f := newFixture()
	challenge := f.challenge.Make(f.request(), _testSlot)

	tests := []struct {
		name           string
		giveDifficulty int
		giveNonce      uint64
		want           bool
	}{
		{"difficulty zero", 0, 42, true},
		{"odd difficulty solved", 3, firstNonce(challenge, 3, true), true},
		{"odd difficulty unsolved", 3, firstNonce(challenge, 3, false), false},
		{"even difficulty solved", 2, firstNonce(challenge, 2, true), true},
		{"even difficulty unsolved", 2, firstNonce(challenge, 2, false), false},
		{"beyond hash length", 65, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newChallenge(tt.giveDifficulty)

			assert.Equal(t, tt.want, c.Check(challenge, tt.giveNonce))
		})
	}
}

func TestChallengeVerify(t *testing.T) {
	const difficulty = 2
	f := newFixture()
	f.challenge = newChallenge(difficulty)
	slot := f.challenge.CurrentSlot()
	current := f.challenge.Make(f.request(), slot)
	previous := f.challenge.Make(f.request(), slot-1)
	stale := f.challenge.Make(f.request(), slot-2)
	other := newFixture()
	other.ip = "198.51.100.1"

	tests := []struct {
		name          string
		giveReq       szhttp.Request
		giveChallenge string
		giveSolved    bool
		want          bool
	}{
		{"current window", f.request(), current, true, true},
		{"previous window", f.request(), previous, true, true},
		{"stale window", f.request(), stale, true, false},
		{"unsolved", f.request(), current, false, false},
		{"other client", other.request(), current, true, false},
		{"forged", f.request(), strings.Repeat("0", 64), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonce := firstNonce(tt.giveChallenge, difficulty, tt.giveSolved)

			got := f.challenge.Verify(tt.giveReq, tt.giveChallenge, nonce)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChallengeSignPass(t *testing.T) {
	f := newFixture()
	exp := time.Now().Add(time.Minute).Unix()

	got := f.challenge.SignPass(f.request(), exp)

	m := _passFmt.FindStringSubmatch(got)
	require.Len(t, m, 3)
	assert.Equal(t, strconv.FormatInt(exp, 10), m[1])
	assert.Equal(t, got, f.challenge.SignPass(f.request(), exp))
	assert.NotEqual(t, f.challenge.Make(f.request(), exp), m[2])
}

func TestChallengeIssuePass(t *testing.T) {
	f := newFixture()

	got := f.challenge.IssuePass(f.request(), time.Hour)

	assert.Equal(t, _passCookie, got.Name)
	assert.Equal(t, 3600, got.MaxAge)
	assert.True(t, got.HttpOnly)

	f.cookie = got.Name + "=" + got.Value
	assert.True(t, f.challenge.ValidPass(f.request()))
}

func TestChallengeValidPass(t *testing.T) {
	f := newFixture()
	now := time.Now().Unix()
	valid := f.challenge.SignPass(f.request(), now+60)
	expired := f.challenge.SignPass(f.request(), now-1)

	tests := []struct {
		name       string
		giveCookie string
		want       bool
	}{
		{"valid pass", _passCookie + "=" + valid, true},
		{"valid among others", "a=b; " + _passCookie + "=" + valid, true},
		{"no cookie header", "", false},
		{"other cookie only", "foo=bar", false},
		{"expired", _passCookie + "=" + expired, false},
		{"tampered signature", _passCookie + "=" + tamper(valid), false},
		{"no separator", _passCookie + "=12345", false},
		{"non-numeric exp", _passCookie + "=abc.deadbeef", false},
		{"non-hex signature", _passCookie + "=9999999999.zz", false},
		{"malformed header", _passCookie, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			f.cookie = tt.giveCookie

			assert.Equal(t, tt.want, f.challenge.ValidPass(f.request()))
		})
	}
}

func TestChallengeValidPassBinding(t *testing.T) {
	f := newFixture()
	exp := time.Now().Add(time.Minute).Unix()
	pass := f.challenge.SignPass(f.request(), exp)

	for _, tt := range bindingTests() {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			tt.giveMutate(f)
			f.cookie = _passCookie + "=" + pass

			assert.Equal(t, tt.wantSame, f.challenge.ValidPass(f.request()))
		})
	}
}

func TestAppendJA4(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{_testJA4, "13d1516_8daaf6152771_02713d6af862"},
		{"q13d0312h3_55b375c5d22e_06cda9e17597",
			"13d0312_55b375c5d22e_06cda9e17597"},
		{"", ""},
		{"q13d", "q13d"},
		{"t13d1516h2x8daaf6152771", "t13d1516h2x8daaf6152771"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, string(appendJA4(nil, tt.give)))
		})
	}
}

type bindingTest struct {
	name       string
	giveMutate func(*fixture)
	wantSame   bool
}

// ja4Of returns a mutation setting the fixture JA4 to ja4.
func ja4Of(ja4 string) func(*fixture) {
	return func(f *fixture) { f.ja4 = ja4 }
}

// bindingTests lists client changes and whether Make and SignPass ignore
// them.
func bindingTests() []bindingTest {
	return []bindingTest{
		{name: "unchanged", giveMutate: func(*fixture) {}, wantSame: true},
		{name: "ip", giveMutate: func(f *fixture) { f.ip = "198.51.100.1" }},
		{name: "user agent", giveMutate: func(f *fixture) { f.ua = "curl" }},
		{
			name:       "conn id",
			giveMutate: func(f *fixture) { f.connId = "other" },
			wantSame:   true,
		},
		{"ja4 transport", ja4Of("q" + _testJA4[1:]), true},
		{"ja4 alpn", ja4Of(_testJA4[:8] + "h1" + _testJA4[10:]), true},
		{"ja4 tls version", ja4Of("t12" + _testJA4[3:]), false},
		{"ja4 ciphers", ja4Of(_testJA4[:11] + "0" + _testJA4[12:]), false},
		{"ja4 malformed", ja4Of("q13d"), false},
		{
			name: "secret",
			giveMutate: func(f *fixture) {
				f.challenge.secret = []byte("other-secret")
			},
		},
	}
}

func newChallenge(difficulty int) *Challenge {
	return New([]byte("test-secret"), difficulty, time.Minute)
}

func header(k, v string) *httppb.RequestHeader {
	return &httppb.RequestHeader{Key: []byte(k), Values: [][]byte{[]byte(v)}}
}

// firstNonce returns the first nonce whose Check result equals want.
func firstNonce(challenge string, difficulty int, want bool) uint64 {
	c := newChallenge(difficulty)
	var nonce uint64
	for c.Check(challenge, nonce) != want {
		nonce++
	}

	return nonce
}

// tamper flips the last hex digit of a pass signature.
func tamper(pass string) string {
	last := pass[len(pass)-1]
	flipped := byte('0')
	if last == '0' {
		flipped = '1'
	}

	return pass[:len(pass)-1] + string(flipped)
}
