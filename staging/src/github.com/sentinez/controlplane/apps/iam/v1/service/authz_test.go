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
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	accrepos "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts"
	accountmock "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts/mock"
	usersrepo "github.com/sentinez/controlplane/apps/iam/v1/repos/users"
	usersmock "github.com/sentinez/controlplane/apps/iam/v1/repos/users/mock"
	membermock "github.com/sentinez/controlplane/apps/tenant/v1/repos/members/mock"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/controlplane/pkg/headers/headerstest"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	_me       = "user_me"
	_coMember = "user_co"
	_stranger = "user_other"
)

func callerCtx(t *testing.T) context.Context {
	t.Helper()

	return headerstest.User(t.Context(), _me)
}

func adminCtx(t *testing.T) context.Context {
	t.Helper()

	return headerstest.Admin(t.Context())
}

// tenantMembers answers the member lookups: _me and _coMember are in
// tnt_1, and _stranger in tnt_2.
func tenantMembers(t *testing.T) *membermock.MockIMember {
	t.Helper()

	all := []*tenantpb.Member{
		{Id: "mem_1", UserId: _me, TenantId: "tnt_1"},
		{Id: "mem_2", UserId: _coMember, TenantId: "tnt_1"},
		{Id: "mem_3", UserId: _stranger, TenantId: "tnt_2"},
	}

	repo := membermock.NewMockIMember(t)
	repo.EXPECT().WithTX(mock.Anything).Return(repo).Maybe()
	repo.EXPECT().List(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, req *tenantpb.ListMemberRequest,
		) (*tenantpb.ListMemberResponse, error) {
			var found []*tenantpb.Member
			for _, m := range all {
				inTenant := len(req.GetTenantIds()) == 0 ||
					req.GetTenantIds()[0] == m.GetTenantId()
				isUser := req.GetUserId() == "" ||
					req.GetUserId() == m.GetUserId()
				if inTenant && isUser {
					found = append(found, m)
				}
			}

			return &tenantpb.ListMemberResponse{Members: found}, nil
		}).Maybe()

	return repo
}

// serviceOf is a service of the tenants of tenantMembers.
func serviceOf(t *testing.T, users usersrepo.IUser, accounts accrepos.IAccount,
) *IAMService {
	t.Helper()

	members := tenantMembers(t)

	return New(nil, nil, nil, users, accounts, members,
		authz.New(members, nil))
}

func TestGetUserVisibility(t *testing.T) {
	tests := []struct {
		name     string
		giveCtx  context.Context
		giveId   string
		wantCode codes.Code
	}{
		{"itself", callerCtx(t), _me, codes.OK},
		{"a user of its tenant", callerCtx(t), _coMember, codes.OK},
		{"a user of another tenant", callerCtx(t), _stranger, codes.NotFound},
		{"any user, as the admin control plane", adminCtx(t), _stranger,
			codes.OK},
		{"nobody signed in", t.Context(), _me, codes.Unauthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := usersmock.NewMockIUser(t)
			users.EXPECT().Get(mock.Anything, tt.giveId).
				Return(&iampb.User{Id: tt.giveId}, nil).Maybe()
			svc := serviceOf(t, users, nil)

			_, err := svc.GetUser(tt.giveCtx, &iampb.GetUserRequest{Id: tt.giveId})

			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

// listedUsers makes the user repo record the request it gets.
func listedUsers(t *testing.T) (*usersmock.MockIUser, *[]*iampb.ListUsersRequest) {
	t.Helper()

	var got []*iampb.ListUsersRequest
	users := usersmock.NewMockIUser(t)
	users.EXPECT().List(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, req *iampb.ListUsersRequest,
		) (*iampb.ListUsersResponse, error) {
			got = append(got, req)
			return &iampb.ListUsersResponse{}, nil
		}).Maybe()

	return users, &got
}

func TestListUsersIsScoped(t *testing.T) {
	tests := []struct {
		name      string
		giveCtx   context.Context
		giveReq   *iampb.ListUsersRequest
		wantIds   []string
		wantEmail []string
	}{
		{"no filter lists only itself", callerCtx(t),
			&iampb.ListUsersRequest{}, []string{_me}, nil},
		{"ids keep the users of its tenants", callerCtx(t),
			&iampb.ListUsersRequest{Ids: []string{_coMember, _stranger}},
			[]string{_coMember}, nil},
		{"a lookup by email goes through", callerCtx(t),
			&iampb.ListUsersRequest{Emails: []string{"new@example.com"}},
			nil, []string{"new@example.com"}},
		{"the admin control plane lists as asked", adminCtx(t),
			&iampb.ListUsersRequest{}, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, got := listedUsers(t)
			svc := serviceOf(t, users, nil)

			_, err := svc.ListUsers(tt.giveCtx, tt.giveReq)

			require.NoError(t, err)
			require.Len(t, *got, 1)
			assert.Equal(t, tt.wantIds, (*got)[0].GetIds())
			assert.Equal(t, tt.wantEmail, (*got)[0].GetEmails())
		})
	}
}

func TestListUsersOfOthersRunsNoQuery(t *testing.T) {
	users, got := listedUsers(t)
	svc := serviceOf(t, users, nil)

	resp, err := svc.ListUsers(callerCtx(t),
		&iampb.ListUsersRequest{Ids: []string{_stranger}})

	require.NoError(t, err)
	assert.Empty(t, resp.GetUsers())
	assert.Empty(t, *got, "an empty filter would list every user")
}

func TestListUsersLooksUpFewEmails(t *testing.T) {
	users, _ := listedUsers(t)
	svc := serviceOf(t, users, nil)

	_, err := svc.ListUsers(callerCtx(t), &iampb.ListUsersRequest{
		Emails: []string{"a@x.io", "b@x.io", "c@x.io", "d@x.io", "e@x.io",
			"f@x.io"},
	})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestListAccounts(t *testing.T) {
	tests := []struct {
		name        string
		giveCtx     context.Context
		giveUserIds []string
		wantUserIds []string
		wantListed  bool
	}{
		{"no filter lists its own", callerCtx(t), nil, []string{_me}, true},
		{"its own id", callerCtx(t), []string{_me}, []string{_me}, true},
		{"the accounts of others", callerCtx(t), []string{_stranger}, nil,
			false},
		{"the admin control plane", adminCtx(t), []string{_stranger},
			[]string{_stranger}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *iampb.ListAccountsRequest
			accounts := accountmock.NewMockIAccount(t)
			accounts.EXPECT().List(mock.Anything, mock.Anything).RunAndReturn(
				func(_ context.Context, req *iampb.ListAccountsRequest,
				) (*iampb.ListAccountsResponse, error) {
					got = req
					return &iampb.ListAccountsResponse{}, nil
				}).Maybe()
			svc := serviceOf(t, nil, accounts)

			_, err := svc.ListAccounts(tt.giveCtx,
				&iampb.ListAccountsRequest{UserIds: tt.giveUserIds})

			require.NoError(t, err)
			assert.Equal(t, tt.wantListed, got != nil)
			assert.Equal(t, tt.wantUserIds, got.GetUserIds())
		})
	}
}

func TestDeleteUserWithoutId(t *testing.T) {
	svc := serviceOf(t, nil, nil)

	_, err := svc.DeleteUser(adminCtx(t), &iampb.DeleteUserRequest{})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestDeleteUserLeavesItsTenants(t *testing.T) {
	conn, err := pgxmock.NewConn()
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close(t.Context()) })
	conn.ExpectBegin()
	conn.ExpectCommit()
	tx, err := conn.Begin(t.Context())
	require.NoError(t, err)

	members := tenantMembers(t)
	members.EXPECT().Delete(mock.Anything, "mem_2").Return(nil).Once()
	accounts := accountmock.NewMockIAccount(t)
	accounts.EXPECT().WithTX(mock.Anything).Return(accounts)
	accounts.EXPECT().DeleteByUserId(mock.Anything, _coMember).Return(nil)
	users := usersmock.NewMockIUser(t)
	users.EXPECT().WithTX(mock.Anything).Return(users)
	users.EXPECT().Delete(mock.Anything, _coMember).Return(nil)
	svc := New(nil, postgres.NewTXMock(tx), nil, users, accounts, members,
		authz.New(members, nil))

	_, err = svc.DeleteUser(adminCtx(t), &iampb.DeleteUserRequest{Id: _coMember})

	require.NoError(t, err)
}
