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

package challenge

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sentinez/core/szcrypto/pow"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/httpconst"
	"github.com/sentinez/shared/bytesconv"
)

// ChallengePath receives the solved proof-of-work form of the challenge
// page.
var ChallengePath = httpconst.SystemPath("challenge")

const (
	PassTTL        = time.Hour
	_powDifficulty = 5
	// _powWindow is how long a challenge page stays solvable: from one to
	// two windows, depending on when in its window the page was made.
	_powWindow  = 5 * time.Minute
	_secretSize = 32
)

// newChallenge signs challenges with the base64 secretKey, or with a random
// key when it is empty: passes then hold only on this process.
func New(secretKey string) (*pow.Challenge, error) {
	secret, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		return nil, fmt.Errorf("decode secret key: %w", err)
	}

	if len(secret) == 0 {
		secret = make([]byte, _secretSize)
		if _, err := rand.Read(secret); err != nil {
			return nil, fmt.Errorf("random secret: %w", err)
		}
	}

	return pow.New(secret, _powDifficulty, _powWindow), nil
}

// IsVerify reports whether ctx posts a solved challenge.
func IsVerify(ctx szhttp.Context) bool {
	return bytesconv.B2s(ctx.Method()) == http.MethodPost &&
		bytesconv.B2s(ctx.Path()) == ChallengePath
}

// localPath returns p when it is a path on this host, else "/", so the
// redirect cannot send the client to another site.
func LocalPath(p string) string {
	if !strings.HasPrefix(p, "/") || strings.ContainsAny(p, "\r\n") {
		return "/"
	}

	if len(p) > 1 && (p[1] == '/' || p[1] == '\\') {
		return "/"
	}

	return p
}
