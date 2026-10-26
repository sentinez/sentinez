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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szhttp"
)

const (
	_passCookie = "pow-pass"

	_tagChallenge = "challenge"
	_tagPass      = "pass"
)

type Challenge struct {
	secret     []byte
	difficulty int
	windowTime time.Duration
}

// New returns a Challenge signed with secret that asks for difficulty zero
// hex digits. A challenge made in one window stays valid through the next.
func New(secret []byte, difficulty int, window time.Duration) *Challenge {
	return &Challenge{
		secret:     append([]byte(nil), secret...),
		difficulty: difficulty,
		windowTime: window,
	}
}

func (c *Challenge) Difficulty() int {
	return c.difficulty
}

func (c *Challenge) CurrentSlot() int64 {
	return time.Now().Unix() / int64(c.windowTime.Seconds())
}

func (c *Challenge) Make(req szhttp.Request, slot int64) string {
	return hex.EncodeToString(c.mac(req, _tagChallenge, slot))
}

// Check reports whether sha256(challenge + nonce) starts with difficulty
// zero hex digits.
func (c *Challenge) Check(challenge string, nonce uint64) bool {
	buf := make([]byte, 0, len(challenge)+20)
	buf = strconv.AppendUint(append(buf, challenge...), nonce, 10)
	sum := sha256.Sum256(buf)

	return zeroNibbles(sum[:], c.difficulty)
}

// Verify reports whether challenge was made for req in the current or the
// previous window and nonce solves it.
func (c *Challenge) Verify(
	req szhttp.Request, challenge string, nonce uint64,
) bool {
	if !c.Check(challenge, nonce) {
		return false
	}

	slot := c.CurrentSlot()
	for _, s := range [...]int64{slot, slot - 1} {
		if hmac.Equal([]byte(c.Make(req, s)), []byte(challenge)) {
			return true
		}
	}

	return false
}

func (c *Challenge) SignPass(req szhttp.Request, exp int64) string {
	buf := append(strconv.AppendInt(nil, exp, 10), '.')

	return string(hex.AppendEncode(buf, c.mac(req, _tagPass, exp)))
}

// IssuePass returns the cookie carrying a pass for req that expires after
// ttl; ValidPass accepts it until then.
func (c *Challenge) IssuePass(
	req szhttp.Request, ttl time.Duration,
) *http.Cookie {
	exp := time.Now().Add(ttl).Unix()

	return &http.Cookie{
		Name:     _passCookie,
		Value:    c.SignPass(req, exp),
		Path:     "/",
		MaxAge:   int(ttl / time.Second),
		Secure:   req.TLS(),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func (c *Challenge) ValidPass(req szhttp.Request) bool {
	value, ok := szhttp.Cookie(req, _passCookie)
	if !ok {
		return false
	}

	expStr, sigHex, ok := strings.Cut(value, ".")
	if !ok {
		return false
	}

	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}

	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}

	return hmac.Equal(sig, c.mac(req, _tagPass, exp))
}

// mac binds tag and n to the client: IP, User-Agent and JA4 (see
// appendJA4). The connection is left out so a pass survives reconnects.
func (c *Challenge) mac(req szhttp.Request, tag string, n int64) []byte {
	buf := make([]byte, 0, 256)
	buf = append(append(buf, tag...), '|')
	buf = append(append(buf, req.RequestIP()...), '|')
	buf = append(append(buf, req.Header(szbytestr.HeaderUserAgent)...), '|')
	buf = append(appendJA4(buf, req.JA4()), '|')
	buf = strconv.AppendInt(buf, n, 10)

	m := hmac.New(sha256.New, c.secret)
	_, _ = m.Write(buf)

	return m.Sum(nil)
}

// appendJA4 appends ja4 without the transport (t, q, d) and ALPN characters
// of its first part, e.g. "13d1516_8daaf6152771_02713d6af862" for
// "t13d1516h2_8daaf6152771_02713d6af862": a browser moving to HTTP/3 or
// another ALPN keeps its pass. Other values are appended whole.
func appendJA4(buf []byte, ja4 string) []byte {
	const alpnEnd = 10 // len("t13d1516h2")
	if len(ja4) <= alpnEnd || ja4[alpnEnd] != '_' {
		return append(buf, ja4...)
	}

	return append(append(buf, ja4[1:alpnEnd-2]...), ja4[alpnEnd:]...)
}

// zeroNibbles reports whether sum starts with n zero hex digits.
func zeroNibbles(sum []byte, n int) bool {
	if n > 2*len(sum) {
		return false
	}

	for _, b := range sum[:n/2] {
		if b != 0 {
			return false
		}
	}

	return n%2 == 0 || sum[n/2]>>4 == 0
}
