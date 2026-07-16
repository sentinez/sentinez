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

package iamsvc

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/sentinez/core/storage/dbx/postgres"
	accrepos "github.com/sentinez/modules/iam/v1/repos/accounts"
	accountrepo "github.com/sentinez/modules/iam/v1/repos/accounts/mock"
	usersrepo "github.com/sentinez/modules/iam/v1/repos/users/mock"
	"github.com/sentinez/modules/pkg/crypto"
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//nolint:funlen
func TestLogin(t *testing.T) {
	ctx := context.Background()

	secBase64 := base64.StdEncoding.EncodeToString([]byte("congchualunglinh"))

	userRepo := usersrepo.NewMockIUser(t)
	accountRepo := accountrepo.NewMockIAccount(t)

	pw, _ := crypto.HashPassword("secret123")
	acc := &accrepos.AccountX{
		Account: &iampb.Account{
			Id:           "acc-123",
			UserId:       "user-123",
			Username:     "admin",
			PasswordHash: pw,
		},
	}
	accountRepo.On("GetByUsernameOrEmail", mock.Anything, "admin").
		Return(acc, nil)

	user := &iampb.User{
		Id:       "user-123",
		FullName: "Test Admin",
	}
	userRepo.On("Get", mock.Anything, "user-123").
		Return(user, nil)

	pgxMock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer func() {
		_ = pgxMock.Close(context.Background())
	}()

	tx, _ := pgxMock.Begin(context.Background())
	txss := postgres.NewTXMock(tx)
	conf := &settingpb.Config{Env: &settingpb.EnvConfig{
		SecretKey: secBase64,
	}}

	config.SetEnv(conf.GetEnv())

	svc := New(conf, txss, nil, userRepo, accountRepo)

	req := &iampb.LoginRequest{
		EmailOrUsername: "admin",
		Password:        "secret123",
	}
	resp, err := svc.Login(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, user, resp.User)
	assert.NotEmpty(t, resp.AccessToken)

	tokenCtx, ok := crypto.BearerTokenVerifier(resp.AccessToken)
	if !ok {
		assert.Error(t, fmt.Errorf("token invalid"))
	}

	assert.True(t, tokenCtx.GetConsole() == typepb.Console_CONSOLE_ADMIN)
}
