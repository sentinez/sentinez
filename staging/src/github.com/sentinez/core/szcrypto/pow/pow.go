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
	"strconv"
	"strings"
	"time"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szcontext"
	"github.com/sentinez/core/szhttp"
)

const (
	_powPass = "pow-pass"

	_tagChallenge = "challenge"
	_tagPass      = "pass"
)

type Challenge struct {
	secret     []byte
	difficulty int
	windowTime time.Duration
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

func (c *Challenge) SignPass(req szhttp.Request, exp int64) string {
	buf := append(strconv.AppendInt(nil, exp, 10), '.')

	return string(hex.AppendEncode(buf, c.mac(req, _tagPass, exp)))
}

func (c *Challenge) ValidPass(req szhttp.Request) bool {
	value, ok := szhttp.Cookie(req, _powPass)
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

// mac binds tag and n to the client: IP, User-Agent, connection and JA4.
func (c *Challenge) mac(req szhttp.Request, tag string, n int64) []byte {
	tr := szcontext.GetTransport(req.Context())

	buf := make([]byte, 0, 256)
	buf = append(append(buf, tag...), '|')
	buf = append(append(buf, req.RequestIP()...), '|')
	buf = append(append(buf, req.Header(szbytestr.HeaderUserAgent)...), '|')
	buf = append(append(buf, tr.GetConnId()...), '|')
	buf = append(append(buf, req.JA4()...), '|')
	buf = strconv.AppendInt(buf, n, 10)

	m := hmac.New(sha256.New, c.secret)
	_, _ = m.Write(buf)

	return m.Sum(nil)
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
