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
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sentinez/pkg/szconfig"
	"github.com/sentinez/pkg/szlog"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	_bearer    = "Bearer "
	_claimsKey = "auth"
	_expKey    = "exp"

	// _issuer signs the tokens, _audience is who they are for: a token the
	// same secret signed for something else is refused.
	_issuer   = "sentinez.iam"
	_audience = "sentinez.controlplane"
)

func TokenGenerator(payload *typepb.Context) (string, error) {
	sec := szconfig.GetEnv(settingpb.Senz_SENZ_SECRET_KEY)
	if sec == "" {
		return "", fmt.Errorf("crypto: secret key is empty")
	}

	return signToken(sec, payload, payload.GetExpireAt().AsTime(),
		_issuer, _audience)
}

func BearerTokenVerifier(bearerToken string) (*typepb.Context, bool) {
	sec := szconfig.GetEnv(settingpb.Senz_SENZ_SECRET_KEY)
	if sec == "" {
		return nil, false
	}

	return verifyToken(bearerToken, sec, _issuer, _audience)
}

// signToken signs payload, valid until exp, with the base64 secret.
func signToken(secretBase64 string, payload *typepb.Context, exp time.Time,
	issuer, audience string) (string, error) {

	secret, err := base64.StdEncoding.DecodeString(secretBase64)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		_claimsKey: payload.String(),
		_expKey:    exp.UTC().Unix(),
		"iss":      issuer,
		"aud":      audience,
		"iat":      time.Now().UTC().Unix(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString(secret)
}

// verifyToken returns the payload of a bearer token signed with the base64
// secret for issuer and audience.
func verifyToken(bearerToken, secretBase64, issuer, audience string,
) (*typepb.Context, bool) {

	token := strings.TrimPrefix(bearerToken, _bearer)
	jwtToken, err := parseJWT(token, secretBase64, issuer, audience)
	if err != nil {
		return nil, false
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return nil, false
	}

	text, ok := claims[_claimsKey].(string)
	if !ok {
		return nil, false
	}

	var resp typepb.Context
	if err := prototext.Unmarshal([]byte(text), &resp); err != nil {
		szlog.Debugf("crypto: error when get claims: %v", err)
		return nil, false
	}

	return &resp, true
}

func parseJWT(token, secretBase64, issuer, audience string,
) (*jwt.Token, error) {
	jwtToken, err := jwt.Parse(token, func(*jwt.Token) (any, error) {
		secret, err := base64.StdEncoding.DecodeString(secretBase64)
		if err != nil {
			return nil, err
		}

		return secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		szlog.Debugf("crypto: invalid token: %v", err)
		return nil, err
	}

	return jwtToken, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
