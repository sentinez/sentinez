// Copyright 2025 Sentinez Labs.
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
	"fmt"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	accountrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/accounts/mock"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users/mock"
	"github.com/sentinez/sentinez/pkg/infra/database/postgres"
	"github.com/sentinez/sentinez/pkg/std/stdcrypto"
	"github.com/sentinez/sentinez/pkg/std/stdperms"
)

//nolint:funlen
func TestLogin(t *testing.T) {
	ctx := context.Background()

	userRepo := usersrepo.NewMockIUser(t)
	accountRepo := accountrepo.NewMockIAccount(t)

	pw, _ := stdcrypto.HashPassword("secret123")
	acc := &iam.Accounts{
		Id:       "acc-123",
		UserId:   "user-123",
		Username: "admin",
		Password: pw,
	}
	accountRepo.On("GetByUsernameOrEmail", mock.Anything, "admin").
		Return(acc, nil)

	user := &iam.Users{
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
	conf := &common.AppConfig{EnvConf: &common.EnvConfig{
		SecretKey: "congchualunglinhlunglinhxinhlunglinh",
	}}

	svc := New(conf, txss, userRepo, accountRepo)

	req := &iam.LoginRequest{
		EmailOrUsername: "admin",
		Password:        "secret123",
	}
	resp, err := svc.Login(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, user, resp.User)
	assert.NotEmpty(t, resp.AccessToken)

	// check permission chứa ROOT
	tokenCtx, ok := stdcrypto.BearerTokenVerifier(
		conf.EnvConf, resp.AccessToken)
	if !ok {
		assert.Error(t, fmt.Errorf("token invalid"))
	}

	assert.True(t, stdperms.Has(
		tokenCtx.PermissionBitwise, common.Permission_PERMISSION_ROOT))
}
