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

	"github.com/golang-jwt/jwt/v5"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/config"
	"github.com/sentinez/shared/zlog"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	bearer    = "Bearer "
	claimsKey = "auth"
	expKey    = "exp"
)

func TokenGenerator(payload *typepb.Context) (string, error) {

	sec := config.Env().GetSecretKey()
	if sec == "" {
		return "", fmt.Errorf("crypto: secret key is empty")
	}

	secret, err := base64.StdEncoding.DecodeString(sec)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		claimsKey: payload.String(),
		expKey:    payload.GetExpireAt().AsTime().UTC().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func BearerTokenVerifier(bearerToken string) (*typepb.Context, bool) {

	sec := config.Env().GetSecretKey()
	if sec == "" {
		return nil, false
	}

	token := strings.TrimPrefix(bearerToken, bearer)
	jwtToken, err := parseJWT(token, sec)
	if err != nil {
		zlog.Debugf("[crypto] invalid token: %v", err)
		return nil, false
	}

	if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok && jwtToken.Valid {
		auth, ok := claims[claimsKey]
		if !ok {
			return nil, false
		}

		var resp typepb.Context
		err = prototext.Unmarshal([]byte(auth.(string)), &resp)
		if err != nil {
			zlog.Debugf("[crypto] error when get claims: %v", err)
			return nil, false
		}

		return &resp, true
	}

	return nil, false
}

func parseJWT(token, secretBase64 string) (*jwt.Token, error) {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil,
				fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		secret, err := base64.StdEncoding.DecodeString(secretBase64)
		if err != nil {
			return "", err
		}

		return secret, nil
	})
	if err != nil {
		zlog.Debugf("[crypto] invalid token: %v", err)
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
