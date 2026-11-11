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
	"strconv"
	"testing"

	accrepos "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts"
	accountmock "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts/mock"
	"github.com/sentinez/pkg/szconfig"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	_adminName = "root.admin"
	_adminPass = "correct horse"
)

// fromAddr is a context of a request the gateway got from addr.
func fromAddr(t *testing.T, forwarded string) context.Context {
	t.Helper()

	return metadata.NewIncomingContext(t.Context(),
		metadata.Pairs("x-forwarded-for", forwarded))
}

func adminService(t *testing.T, password string) *IAMService {
	t.Helper()

	secret := base64.StdEncoding.EncodeToString([]byte("congchualunglinh"))
	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY, secret)

	return New(&settingpb.Config{Env: map[string]string{
		settingpb.Senz_SENZ_ADMIN_USERNAME.String(): _adminName,
		settingpb.Senz_SENZ_ADMIN_PASSWORD.String(): password,
	}}, nil, nil, nil, nil, nil, nil)
}

func loginAs(username, password string) *iampb.LoginRequest {
	return &iampb.LoginRequest{EmailOrUsername: username, Password: password}
}

func TestAdminLogin(t *testing.T) {
	tests := []struct {
		name         string
		giveConfig   string
		giveInput    string
		wantCode     codes.Code
		wantAdminJwt bool
	}{
		{"right password", _adminPass, _adminPass, codes.OK, true},
		{"wrong password", _adminPass, "wrong password", codes.Unauthenticated,
			false},
		{"a prefix of it", _adminPass, "correct", codes.Unauthenticated,
			false},
		{"none configured, none given", "", "", codes.Unauthenticated, false},
		{"none configured, any given", "", "anything", codes.Unauthenticated,
			false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := adminService(t, tt.giveConfig)

			resp, err := svc.Login(t.Context(),
				loginAs(_adminName, tt.giveInput))

			assert.Equal(t, tt.wantCode, status.Code(err))
			assert.Equal(t, tt.wantAdminJwt, resp.GetAccessToken() != "")
		})
	}
}

// wrongPasswords fails n logins of username from the address.
func wrongPasswords(t *testing.T, svc *IAMService, addr, username string, n int,
) {
	t.Helper()

	for range n {
		_, err := svc.Login(fromAddr(t, addr), loginAs(username, "wrong one"))
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	}
}

func TestLoginBlocksAfterFailures(t *testing.T) {
	svc := adminService(t, _adminPass)

	wrongPasswords(t, svc, "10.0.0.1", _adminName, _accountFailures)

	// even the right password: guessing is stopped, not only slowed
	_, err := svc.Login(fromAddr(t, "10.0.0.1"),
		loginAs(_adminName, _adminPass))
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}

func TestLoginBlockIsPerClient(t *testing.T) {
	svc := adminService(t, _adminPass)
	wrongPasswords(t, svc, "10.0.0.1", _adminName, _accountFailures)

	_, err := svc.Login(fromAddr(t, "10.0.0.2"),
		loginAs(_adminName, _adminPass))

	assert.NoError(t, err)
}

func TestLoginBlockIsPerAccount(t *testing.T) {
	svc := adminService(t, _adminPass)
	wrongPasswords(t, svc, "10.0.0.1", _adminName, _accountFailures)

	// another account from the same client is not locked by that
	other := attemptOf(fromAddr(t, "10.0.0.1"), "someone.else")

	assert.NoError(t, svc.limits.check(other))
}

func TestLoginSuccessForgetsTheAccountFailures(t *testing.T) {
	svc := adminService(t, _adminPass)
	ctx := fromAddr(t, "10.0.0.1")

	wrongPasswords(t, svc, "10.0.0.1", _adminName, _accountFailures-1)
	_, err := svc.Login(ctx, loginAs(_adminName, _adminPass))
	require.NoError(t, err)

	wrongPasswords(t, svc, "10.0.0.1", _adminName, _accountFailures-1)
}

func TestLoginBlocksAClientSprayingAccounts(t *testing.T) {
	svc := adminService(t, _adminPass)

	for i := range _clientFailures {
		svc.limits.fail(attemptOf(fromAddr(t, "10.0.0.1"),
			"user"+strconv.Itoa(i)))
	}

	_, err := svc.Login(fromAddr(t, "10.0.0.1"), loginAs("fresh.name", "x"))
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}

func TestLoginCanNotEscapeBySpoofingForwardedFor(t *testing.T) {
	svc := adminService(t, _adminPass)

	// the client writes the first hops; the last is what the gateway saw
	for i := range _accountFailures {
		forged := "1.1.1." + string(rune('0'+i)) + ", 10.0.0.1"
		_, _ = svc.Login(fromAddr(t, forged), loginAs(_adminName, "wrong one"))
	}

	_, err := svc.Login(fromAddr(t, "7.7.7.7, 10.0.0.1"),
		loginAs(_adminName, _adminPass))

	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}

func TestClientAddr(t *testing.T) {
	tests := []struct {
		name string
		give context.Context
		want string
	}{
		{"one hop", metadata.NewIncomingContext(t.Context(),
			metadata.Pairs("x-forwarded-for", "10.0.0.1")), "10.0.0.1"},
		{"the last of several", metadata.NewIncomingContext(t.Context(),
			metadata.Pairs("x-forwarded-for", "6.6.6.6, 10.0.0.1")),
			"10.0.0.1"},
		{"nothing known", t.Context(), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, clientAddr(tt.give))
		})
	}
}

func TestLocalAccountsFailureCounts(t *testing.T) {
	accounts := accountmock.NewMockIAccount(t)
	accounts.EXPECT().GetByProvider(mock.Anything, mock.Anything,
		mock.Anything).Return((*accrepos.AccountX)(nil), assert.AnError)
	secret := base64.StdEncoding.EncodeToString([]byte("congchualunglinh"))
	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY, secret)
	svc := New(&settingpb.Config{}, nil, nil, nil, accounts, nil, nil)

	for range _accountFailures {
		_, err := svc.Login(fromAddr(t, "10.0.0.1"),
			loginAs("nobody@example.com", "whatever1"))
		require.Error(t, err)
	}

	_, err := svc.Login(fromAddr(t, "10.0.0.1"),
		loginAs("nobody@example.com", "whatever1"))
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}
