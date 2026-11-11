// Copyright 2025 Duc-Hung Ho.
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

package crypto

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sentinez/pkg/szconfig"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGenAndVerifyToken(t *testing.T) {
	secBase64 := base64.StdEncoding.EncodeToString([]byte("congchualunglinh"))

	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY, secBase64)

	token, err := TokenGenerator(&typepb.Context{
		Name:     "test gen & verify",
		ExpireAt: timestamppb.New(time.Now().Add(time.Hour)),
	})
	if err != nil {
		t.Error(err)
	}

	tp, ok := BearerTokenVerifier(token)
	if !ok {
		t.Error("fail to verify bearer token")
	}

	t.Log("token payload: ", tp.String())
}

const _testSecret = "congchualunglinh"

// signed returns a token of claims signed with method and secret.
func signed(t *testing.T, method jwt.SigningMethod, secret any,
	claims jwt.MapClaims) string {
	t.Helper()

	token, err := jwt.NewWithClaims(method, claims).SignedString(secret)
	require.NoError(t, err)

	return token
}

// claimsWith are the claims of a valid token, with the changes of set.
func claimsWith(set func(jwt.MapClaims)) jwt.MapClaims {
	claims := jwt.MapClaims{
		_claimsKey: (&typepb.Context{UserId: "user_1"}).String(),
		_expKey:    time.Now().Add(time.Hour).Unix(),
		"iss":      _issuer,
		"aud":      _audience,
	}
	set(claims)

	return claims
}

func TestBearerTokenVerifier_Accepts(t *testing.T) {
	secret := []byte(_testSecret)
	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY,
		base64.StdEncoding.EncodeToString(secret))

	token := signed(t, jwt.SigningMethodHS256, secret,
		claimsWith(func(jwt.MapClaims) {}))

	got, ok := BearerTokenVerifier(token)

	require.True(t, ok)
	assert.Equal(t, "user_1", got.GetUserId())
}

//nolint:funlen
func TestBearerTokenVerifier_Refuses(t *testing.T) {
	secret := []byte(_testSecret)
	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY,
		base64.StdEncoding.EncodeToString(secret))

	tests := []struct {
		name string
		give string
	}{
		{"another audience", signed(t, jwt.SigningMethodHS256, secret,
			claimsWith(func(c jwt.MapClaims) { c["aud"] = "sentinez.edge" }))},
		{"no audience", signed(t, jwt.SigningMethodHS256, secret,
			claimsWith(func(c jwt.MapClaims) { delete(c, "aud") }))},
		{"another issuer", signed(t, jwt.SigningMethodHS256, secret,
			claimsWith(func(c jwt.MapClaims) { c["iss"] = "someone" }))},
		{"expired", signed(t, jwt.SigningMethodHS256, secret,
			claimsWith(func(c jwt.MapClaims) {
				c[_expKey] = time.Now().Add(-time.Hour).Unix()
			}))},
		{"no expiry", signed(t, jwt.SigningMethodHS256, secret,
			claimsWith(func(c jwt.MapClaims) { delete(c, _expKey) }))},
		{"another secret", signed(t, jwt.SigningMethodHS256,
			[]byte("another secret"), claimsWith(func(jwt.MapClaims) {}))},
		{"another HMAC algorithm", signed(t, jwt.SigningMethodHS512, secret,
			claimsWith(func(jwt.MapClaims) {}))},
		{"no signature", signed(t, jwt.SigningMethodNone,
			jwt.UnsafeAllowNoneSignatureType,
			claimsWith(func(jwt.MapClaims) {}))},
		{"not a token", "not-a-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := BearerTokenVerifier(tt.give)

			assert.False(t, ok)
		})
	}
}

func TestTokenGenerator_Claims(t *testing.T) {
	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY,
		base64.StdEncoding.EncodeToString([]byte(_testSecret)))

	token, err := TokenGenerator(&typepb.Context{
		UserId:   "user_1",
		ExpireAt: timestamppb.New(time.Now().Add(time.Hour)),
	})
	require.NoError(t, err)

	parsed, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	require.NoError(t, err)

	iss, _ := parsed.Claims.GetIssuer()
	aud, _ := parsed.Claims.GetAudience()
	iat, _ := parsed.Claims.GetIssuedAt()
	assert.Equal(t, _issuer, iss)
	assert.Equal(t, jwt.ClaimStrings{_audience}, aud)
	assert.NotNil(t, iat)
}
