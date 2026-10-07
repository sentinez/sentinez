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
	"errors"
	"testing"

	accrepos "github.com/sentinez/controlplane/iam/v1/repos/accounts"
	accountrepo "github.com/sentinez/controlplane/iam/v1/repos/accounts/mock"
	usersrepo "github.com/sentinez/controlplane/iam/v1/repos/users/mock"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/errorx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errDB = errors.New("db down")

type lookup struct {
	acc *accrepos.AccountX
	err error
}

func existing(email string) lookup {
	return lookup{acc: &accrepos.AccountX{Account: &iampb.Account{
		Id: "senz.iam.accounts.1", Email: email, UserId: "senz.iam.users.1",
	}}}
}

var missing = lookup{acc: &accrepos.AccountX{}, err: errorx.ErrNotFound}

func newTestService(t *testing.T,
	lookups map[string]lookup) (*IAMService, *usersrepo.MockIUser) {
	t.Helper()
	userRepo := usersrepo.NewMockIUser(t)
	accountRepo := accountrepo.NewMockIAccount(t)
	for input, l := range lookups {
		accountRepo.On("GetByUsernameOrEmail", mock.Anything, input).
			Return(l.acc, l.err)
	}
	svc := New(&settingpb.Config{}, nil, nil, userRepo, accountRepo)
	return svc, userRepo
}

func TestUsernameOrEmailMustUnique(t *testing.T) {
	tests := []struct {
		name         string
		giveUsername string
		giveEmail    string
		giveLookups  map[string]lookup
		wantCode     codes.Code
		wantErr      error
	}{
		{
			name:         "both free",
			giveUsername: "alice",
			giveEmail:    "alice@x.io",
			giveLookups: map[string]lookup{
				"alice": missing, "alice@x.io": missing,
			},
			wantCode: codes.OK,
		},
		{
			name:         "username taken",
			giveUsername: "alice",
			giveEmail:    "alice@x.io",
			giveLookups:  map[string]lookup{"alice": existing("a@x.io")},
			wantCode:     codes.AlreadyExists,
		},
		{
			name:         "email taken",
			giveUsername: "alice",
			giveEmail:    "alice@x.io",
			giveLookups: map[string]lookup{
				"alice": missing, "alice@x.io": existing("alice@x.io"),
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name:      "empty username only checks email",
			giveEmail: "alice@x.io",
			giveLookups: map[string]lookup{
				"alice@x.io": existing("alice@x.io"),
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name:         "lookup failure is returned",
			giveUsername: "alice",
			giveEmail:    "alice@x.io",
			giveLookups:  map[string]lookup{"alice": {err: errDB}},
			wantErr:      errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newTestService(t, tt.giveLookups)

			err := svc.UsernameOrEmailMustUnique(
				context.Background(), tt.giveUsername, tt.giveEmail)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestCreateUser_EmailTaken(t *testing.T) {
	svc, _ := newTestService(t,
		map[string]lookup{"bob@x.io": existing("bob@x.io")})

	resp, err := svc.CreateUser(context.Background(),
		&iampb.CreateUserRequest{Email: "bob@x.io", FullName: "Bob"})

	assert.Nil(t, resp)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestCreateUser_EmailFree(t *testing.T) {
	svc, userRepo := newTestService(t,
		map[string]lookup{"bob@x.io": missing})
	userRepo.On("Create", mock.Anything, mock.Anything).
		Return(&iampb.User{Id: "senz.iam.users.2"}, nil)

	resp, err := svc.CreateUser(context.Background(),
		&iampb.CreateUserRequest{Email: "bob@x.io", FullName: "Bob"})

	require.NoError(t, err)
	assert.Equal(t, "senz.iam.users.2", resp.GetUserId())
}

func TestUpdateUser_AccountMissing(t *testing.T) {
	svc, _ := newTestService(t, map[string]lookup{"bob@x.io": missing})

	resp, err := svc.UpdateUser(context.Background(),
		&iampb.UpdateUserRequest{Id: "senz.iam.users.2", Email: "bob@x.io"})

	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestUpdateUser_AccountExists(t *testing.T) {
	svc, userRepo := newTestService(t,
		map[string]lookup{"bob@x.io": existing("bob@x.io")})
	user := &iampb.User{Id: "senz.iam.users.2", FullName: "Bob"}
	userRepo.On("Get", mock.Anything, "senz.iam.users.2").Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.MatchedBy(
		func(u *iampb.User) bool { return u.GetFullName() == "Robert" },
	)).Return(nil)

	resp, err := svc.UpdateUser(context.Background(), &iampb.UpdateUserRequest{
		Id: "senz.iam.users.2", Email: "bob@x.io", FullName: "Robert",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
}
