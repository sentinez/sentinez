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

	accrepos "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts"
	accountrepo "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts/mock"
	usersrepo "github.com/sentinez/controlplane/apps/iam/v1/repos/users/mock"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szerror"
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
		Id: "acct_1", Email: email, UserId: "user_1",
	}}}
}

var _missing = lookup{acc: &accrepos.AccountX{}, err: szerror.ErrNotFound}

// newTestService returns a service whose account repo answers lookups of
// any provider (GetByUsernameOrEmail) with lookups.
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

// newLocalTestService is newTestService for lookups limited to the local
// provider (GetByProvider), with "root" as the reserved admin username.
func newLocalTestService(t *testing.T,
	lookups map[string]lookup) (*IAMService, *usersrepo.MockIUser) {
	t.Helper()
	userRepo := usersrepo.NewMockIUser(t)
	accountRepo := accountrepo.NewMockIAccount(t)
	for input, l := range lookups {
		accountRepo.On("GetByProvider",
			mock.Anything, iampb.Provider_PROVIDER_LOCAL, input).
			Return(l.acc, l.err)
	}
	conf := &settingpb.Config{Env: map[string]string{
		settingpb.Senz_SENZ_ADMIN_USERNAME.String(): "root",
	}}
	svc := New(conf, nil, nil, userRepo, accountRepo)
	return svc, userRepo
}

// nolint
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
				"alice": _missing, "alice@x.io": _missing,
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
				"alice": _missing, "alice@x.io": existing("alice@x.io"),
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
			name:         "admin username is reserved",
			giveUsername: "root",
			giveEmail:    "alice@x.io",
			wantCode:     codes.AlreadyExists,
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
			svc, _ := newLocalTestService(t, tt.giveLookups)

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

func TestCreateAccount_EmailOwnedByUser(t *testing.T) {
	// no local account uses the email, but a passkey user owns it
	svc, userRepo := newLocalTestService(t, map[string]lookup{
		"bob": _missing, "bob@x.io": _missing,
	})
	userRepo.On("GetByEmail", mock.Anything, "bob@x.io").
		Return(&iampb.User{Id: "user_1"}, nil)

	resp, err := svc.CreateAccount(context.Background(),
		&iampb.CreateAccountRequest{
			Username: "bob", Email: "bob@x.io", Password: "secret123",
		})

	assert.Nil(t, resp)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestCreateUser_EmailTaken(t *testing.T) {
	svc, userRepo := newTestService(t, nil)
	userRepo.On("GetByEmail", mock.Anything, "bob@x.io").
		Return(&iampb.User{Id: "user_1"}, nil)

	resp, err := svc.CreateUser(context.Background(),
		&iampb.CreateUserRequest{Email: "bob@x.io", FullName: "Bob"})

	assert.Nil(t, resp)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestCreateUser_EmailFree(t *testing.T) {
	svc, userRepo := newTestService(t, nil)
	userRepo.On("GetByEmail", mock.Anything, "bob@x.io").
		Return(nil, szerror.ErrNotFound)
	userRepo.On("Create", mock.Anything, mock.Anything).
		Return(&iampb.User{Id: "user_2"}, nil)

	resp, err := svc.CreateUser(context.Background(),
		&iampb.CreateUserRequest{Email: "bob@x.io", FullName: "Bob"})

	require.NoError(t, err)
	assert.Equal(t, "user_2", resp.GetUserId())
}

func TestUpdateUser_RenamesSignedInUser(t *testing.T) {
	svc, userRepo := newTestService(t, nil)
	user := &iampb.User{
		Id: "user_2", FullName: "Bob", Email: "bob@x.io",
	}
	userRepo.On("Get", mock.Anything, "user_2").Return(user, nil)
	// only the id and the name reach the store
	userRepo.On("Update", mock.Anything, mock.MatchedBy(
		func(u *iampb.User) bool {
			return u.GetId() == "user_2" &&
				u.GetFullName() == "Robert" && u.GetEmail() == ""
		},
	)).Return(nil)

	resp, err := svc.UpdateUser(signedIn(t, "user_2"),
		&iampb.UpdateUserRequest{FullName: " Robert "})

	require.NoError(t, err)
	assert.Equal(t, "Robert", resp.GetUser().GetFullName())
	assert.Equal(t, "bob@x.io", resp.GetUser().GetEmail())
}

func TestUpdateUser_Rejected(t *testing.T) {
	tests := []struct {
		name     string
		giveCtx  context.Context
		giveName string
		wantCode codes.Code
	}{
		{
			name:     "not signed in",
			giveCtx:  context.Background(),
			giveName: "Robert",
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "blank name",
			giveCtx:  signedIn(t, "user_2"),
			giveName: "  ",
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newTestService(t, nil)

			resp, err := svc.UpdateUser(tt.giveCtx,
				&iampb.UpdateUserRequest{FullName: tt.giveName})

			assert.Nil(t, resp)
			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}
