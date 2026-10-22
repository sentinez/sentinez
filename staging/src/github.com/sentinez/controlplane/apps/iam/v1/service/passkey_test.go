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

package iamsvc

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	accrepos "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts"
	accountrepo "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts/mock"
	usersrepo "github.com/sentinez/controlplane/apps/iam/v1/repos/users/mock"
	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/controlplane/pkg/headers"
	"github.com/sentinez/controlplane/pkg/passkey"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szconfig"
	"github.com/sentinez/shared/szerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	_bobId    = "senz.iam.users.bob"
	_bobEmail = "bob@x.io"
)

type passkeyFixture struct {
	svc      *IAMService
	store    passkey.Store
	users    *usersrepo.MockIUser
	accounts *accountrepo.MockIAccount
}

// newPasskeyFixture returns a service with a working WebAuthn relying
// party. The mocks fail the test on any call that is not expected, so a
// test that sets no Create expectation proves nothing was persisted.
func newPasskeyFixture(t *testing.T) *passkeyFixture {
	t.Helper()

	f := &passkeyFixture{
		store:    passkey.NewMemoryStorage(time.Minute),
		users:    usersrepo.NewMockIUser(t),
		accounts: accountrepo.NewMockIAccount(t),
	}
	conf := &settingpb.Config{Env: map[string]string{
		settingpb.Senz_SENZ_HOSTNAME.String():      "localhost",
		settingpb.Senz_SENZ_CLIENT_ORIGIN.String(): "http://localhost",
	}}
	f.svc = New(conf, nil, f.store, f.users, f.accounts)

	return f
}

// signedIn returns a context carrying an access token of userId.
func signedIn(t *testing.T, userId string) context.Context {
	t.Helper()

	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY,
		base64.StdEncoding.EncodeToString([]byte("congchualunglinh")))

	token, err := crypto.TokenGenerator(&typepb.Context{
		UserId:       userId,
		ExpireAt:     timestamppb.New(time.Now().Add(time.Hour)),
		ControlPlane: typepb.ControlPlane_CONTROL_PLANE_PORTAL,
	})
	require.NoError(t, err)

	return metadata.NewIncomingContext(context.Background(),
		metadata.Pairs(headers.AuthHeader, token))
}

func passkeyAccountOf(userId, handle string,
	credentials ...string) *accrepos.AccountX {
	return &accrepos.AccountX{Account: &iampb.Account{
		Id:             "senz.iam.accounts.pk",
		UserId:         userId,
		Provider:       iampb.Provider_PROVIDER_PASSKEY,
		ProviderUserId: handle,
		Credentials:    credentials,
	}}
}

func TestPasskeyRegisterChallenge_Rejected(t *testing.T) {
	tests := []struct {
		name     string
		give     string
		wantCode codes.Code
	}{
		{name: "username", give: "bob", wantCode: codes.InvalidArgument},
		{name: "email owned", give: _bobEmail, wantCode: codes.AlreadyExists},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPasskeyFixture(t)
			f.users.On("GetByEmail", mock.Anything, _bobEmail).
				Return(&iampb.User{Id: _bobId}, nil).Maybe()

			resp, err := f.svc.PasskeyRegisterChallenge(context.Background(),
				&iampb.PasskeyRegisterChallengeRequest{
					EmailOrUsername: tt.give,
				})

			assert.Nil(t, resp)
			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestPasskeyRegisterChallenge_PersistsNothing(t *testing.T) {
	f := newPasskeyFixture(t)
	f.users.On("GetByEmail", mock.Anything, _bobEmail).
		Return(nil, szerror.ErrNotFound)

	resp, err := f.svc.PasskeyRegisterChallenge(context.Background(),
		&iampb.PasskeyRegisterChallengeRequest{EmailOrUsername: _bobEmail})
	require.NoError(t, err)

	reg, ok := f.store.TakeRegistration(resp.GetSessionId())
	require.True(t, ok)
	assert.Empty(t, reg.UserId)
	assert.Equal(t, _bobEmail, reg.Email)
	assert.NotEmpty(t, reg.Handle)
	assert.NotEqual(t, _bobEmail, reg.Handle)
}

func TestPasskeyRegisterVerify_UnknownSession(t *testing.T) {
	f := newPasskeyFixture(t)

	resp, err := f.svc.PasskeyRegisterVerify(context.Background(),
		&iampb.PasskeyRegisterVerifyRequest{SessionId: "nope"})

	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// A response that fails verification must leave no user or account
// behind, and must not be retryable on the same session.
func TestPasskeyRegisterVerify_BadResponse(t *testing.T) {
	f := newPasskeyFixture(t)
	f.users.On("GetByEmail", mock.Anything, _bobEmail).
		Return(nil, szerror.ErrNotFound)

	ctx := context.Background()
	challenge, err := f.svc.PasskeyRegisterChallenge(ctx,
		&iampb.PasskeyRegisterChallengeRequest{EmailOrUsername: _bobEmail})
	require.NoError(t, err)

	req := &iampb.PasskeyRegisterVerifyRequest{
		SessionId:                  challenge.GetSessionId(),
		CredentialCreationResponse: []byte(`{"id":"forged"}`),
	}

	_, err = f.svc.PasskeyRegisterVerify(ctx, req)
	require.Error(t, err)

	_, err = f.svc.PasskeyRegisterVerify(ctx, req)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestPasskeyLinkChallenge_Unauthenticated(t *testing.T) {
	f := newPasskeyFixture(t)

	resp, err := f.svc.PasskeyLinkChallenge(
		context.Background(), &iampb.PasskeyLinkChallengeRequest{})

	assert.Nil(t, resp)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestPasskeyLinkChallenge_FirstPasskey(t *testing.T) {
	f := newPasskeyFixture(t)
	f.users.On("Get", mock.Anything, _bobId).
		Return(&iampb.User{Id: _bobId, Email: _bobEmail}, nil)
	f.accounts.On("GetByUserId",
		mock.Anything, iampb.Provider_PROVIDER_PASSKEY, _bobId).
		Return(&accrepos.AccountX{}, szerror.ErrNotFound)

	resp, err := f.svc.PasskeyLinkChallenge(
		signedIn(t, _bobId), &iampb.PasskeyLinkChallengeRequest{})
	require.NoError(t, err)

	reg, ok := f.store.TakeRegistration(resp.GetSessionId())
	require.True(t, ok)
	assert.Equal(t, _bobId, reg.UserId)
	assert.Equal(t, _bobEmail, reg.Email)
	assert.NotEmpty(t, reg.Handle)
}

func TestPasskeyLinkChallenge_ReusesHandle(t *testing.T) {
	f := newPasskeyFixture(t)
	f.users.On("Get", mock.Anything, _bobId).
		Return(&iampb.User{Id: _bobId, Email: _bobEmail}, nil)
	f.accounts.On("GetByUserId",
		mock.Anything, iampb.Provider_PROVIDER_PASSKEY, _bobId).
		Return(passkeyAccountOf(_bobId, "handle-1"), nil)

	resp, err := f.svc.PasskeyLinkChallenge(
		signedIn(t, _bobId), &iampb.PasskeyLinkChallengeRequest{})
	require.NoError(t, err)

	reg, ok := f.store.TakeRegistration(resp.GetSessionId())
	require.True(t, ok)
	assert.Equal(t, "handle-1", reg.Handle)
}

// nolint:funlen
func TestAuthorizeLink(t *testing.T) {
	tests := []struct {
		name     string
		giveCtx  context.Context
		giveReg  *passkey.Registration
		wantCode codes.Code
	}{
		{
			name:     "sign up needs no token",
			giveCtx:  context.Background(),
			giveReg:  &passkey.Registration{Email: _bobEmail},
			wantCode: codes.OK,
		},
		{
			name:     "link without token",
			giveCtx:  context.Background(),
			giveReg:  &passkey.Registration{UserId: _bobId},
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "link by another user",
			giveCtx:  signedIn(t, "senz.iam.users.eve"),
			giveReg:  &passkey.Registration{UserId: _bobId},
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "link by the owner",
			giveCtx:  signedIn(t, _bobId),
			giveReg:  &passkey.Registration{UserId: _bobId},
			wantCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authorizeLink(tt.giveCtx, tt.giveReg)

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestSavePasskey_LinkFirstPasskey(t *testing.T) {
	f := newPasskeyFixture(t)
	f.accounts.On("GetByUserId",
		mock.Anything, iampb.Provider_PROVIDER_PASSKEY, _bobId).
		Return(&accrepos.AccountX{}, szerror.ErrNotFound)
	f.accounts.On("Create", mock.Anything, mock.MatchedBy(
		func(acc *accrepos.AccountX) bool {
			return acc.GetUserId() == _bobId &&
				acc.GetProvider() == iampb.Provider_PROVIDER_PASSKEY &&
				acc.GetProviderUserId() == "handle-1" &&
				acc.GetPasswordHash() == "" &&
				len(acc.GetCredentials()) == 1
		},
	)).Return(passkeyAccountOf(_bobId, "handle-1"), nil)

	err := f.svc.savePasskey(context.Background(), &passkey.Registration{
		UserId: _bobId, Email: _bobEmail, Handle: "handle-1",
	}, &webauthn.Credential{ID: []byte("cred-1")})

	require.NoError(t, err)
}

func TestSavePasskey_LinkAnotherPasskey(t *testing.T) {
	f := newPasskeyFixture(t)
	f.accounts.On("GetByUserId",
		mock.Anything, iampb.Provider_PROVIDER_PASSKEY, _bobId).
		Return(passkeyAccountOf(_bobId, "handle-1", `{"id":"Y3JlZC0x"}`), nil)
	f.accounts.On("Update", mock.Anything, mock.MatchedBy(
		func(acc *accrepos.AccountX) bool {
			return len(acc.GetCredentials()) == 2
		},
	)).Return(nil)

	err := f.svc.savePasskey(context.Background(), &passkey.Registration{
		UserId: _bobId, Email: _bobEmail, Handle: "handle-1",
	}, &webauthn.Credential{ID: []byte("cred-2")})

	require.NoError(t, err)
}

func TestSavePasskey_StaleHandle(t *testing.T) {
	f := newPasskeyFixture(t)
	f.accounts.On("GetByUserId",
		mock.Anything, iampb.Provider_PROVIDER_PASSKEY, _bobId).
		Return(passkeyAccountOf(_bobId, "handle-1"), nil)

	err := f.svc.savePasskey(context.Background(), &passkey.Registration{
		UserId: _bobId, Email: _bobEmail, Handle: "handle-2",
	}, &webauthn.Credential{ID: []byte("cred-2")})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSavePasskey_SignUpEmailTaken(t *testing.T) {
	f := newPasskeyFixture(t)
	f.users.On("GetByEmail", mock.Anything, _bobEmail).
		Return(&iampb.User{Id: _bobId}, nil)

	err := f.svc.savePasskey(context.Background(), &passkey.Registration{
		Email: _bobEmail, Handle: "handle-1",
	}, &webauthn.Credential{ID: []byte("cred-1")})

	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestPasskeyLoginChallenge_NoPasskeyAccount(t *testing.T) {
	f := newPasskeyFixture(t)
	f.accounts.On("GetByProvider",
		mock.Anything, iampb.Provider_PROVIDER_PASSKEY, _bobEmail).
		Return(&accrepos.AccountX{}, szerror.ErrNotFound)

	resp, err := f.svc.PasskeyLoginChallenge(context.Background(),
		&iampb.PasskeyLoginChallengeRequest{EmailOrUsername: _bobEmail})

	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestPasskeyLoginVerify_LooksUpByHandle(t *testing.T) {
	f := newPasskeyFixture(t)
	f.store.SaveSession("sid", &webauthn.SessionData{
		UserID: []byte("handle-1"),
	})
	f.accounts.On("GetByProviderUserId",
		mock.Anything, iampb.Provider_PROVIDER_PASSKEY, "handle-1").
		Return(&accrepos.AccountX{}, szerror.ErrNotFound)

	resp, err := f.svc.PasskeyLoginVerify(context.Background(),
		&iampb.PasskeyLoginVerifyRequest{SessionId: "sid"})

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, szerror.ErrNotFound)
}
