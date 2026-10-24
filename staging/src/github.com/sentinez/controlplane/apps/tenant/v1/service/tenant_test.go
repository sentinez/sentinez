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

package tenantsvc

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	membermock "github.com/sentinez/controlplane/apps/tenant/v1/repos/members/mock"
	tenantmock "github.com/sentinez/controlplane/apps/tenant/v1/repos/tenants/mock"
	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/controlplane/pkg/headers"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szconfig"
	"github.com/sentinez/shared/szstorage/dbx/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _createTenantReq = &tenantpb.CreateTenantRequest{
	Name:        "acme",
	Description: "acme team",
	Plan:        typepb.Plan_PLAN_FREE,
}

// authContext returns a context carrying a bearer token of userId.
func authContext(t *testing.T, userId string) context.Context {
	t.Helper()

	secret := base64.StdEncoding.EncodeToString([]byte("congchualunglinh"))
	szconfig.SetEnv(settingpb.Senz_SENZ_SECRET_KEY, secret)

	token, err := crypto.TokenGenerator(&typepb.Context{
		UserId:   userId,
		ExpireAt: timestamppb.New(time.Now().Add(time.Hour)),
	})
	require.NoError(t, err)

	md := metadata.Pairs(headers.AuthHeader, "Bearer "+token)
	return metadata.NewIncomingContext(t.Context(), md)
}

// mockTx returns a Tx backed by pgxmock; the caller sets what must
// happen to the transaction (commit or rollback).
func mockTx(t *testing.T) (*postgres.Tx, pgxmock.PgxConnIface) {
	t.Helper()

	conn, err := pgxmock.NewConn()
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close(context.Background()) })

	conn.ExpectBegin()
	tx, err := conn.Begin(t.Context())
	require.NoError(t, err)

	return postgres.NewTXMock(tx), conn
}

// tenantRepo expects the tenant of _createTenantReq to be created.
func tenantRepo(t *testing.T) *tenantmock.MockITenant {
	t.Helper()

	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().WithTX(mock.Anything).Return(repo)
	repo.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context,
			tenant *tenantpb.Tenant) (*tenantpb.Tenant, error) {
			tenant.Id = _tenantId
			return tenant, nil
		})

	return repo
}

func TestCreateTenant(t *testing.T) {
	tx, conn := mockTx(t)
	conn.ExpectCommit()

	members := membermock.NewMockIMember(t)
	members.EXPECT().WithTX(mock.Anything).Return(members)
	members.EXPECT().Create(mock.Anything, &tenantpb.Member{
		UserId:   _userId,
		TenantId: _tenantId,
		Roles:    []typepb.Role{typepb.Role_ROLE_ADMIN},
	}).Return(&tenantpb.Member{}, nil)

	resp, err := New(tx, nil, members, tenantRepo(t)).
		CreateTenant(authContext(t, _userId), _createTenantReq)

	require.NoError(t, err)
	assert.Equal(t, _tenantId, resp.GetTenant().GetId())
	assert.Equal(t, "acme", resp.GetTenant().GetName())
	assert.NoError(t, conn.ExpectationsWereMet())
}

func TestCreateTenantRollsBackWhenMemberFails(t *testing.T) {
	tx, conn := mockTx(t)
	conn.ExpectRollback()

	members := membermock.NewMockIMember(t)
	members.EXPECT().WithTX(mock.Anything).Return(members)
	members.EXPECT().Create(mock.Anything, mock.Anything).
		Return(nil, errors.New("insert member"))

	_, err := New(tx, nil, members, tenantRepo(t)).
		CreateTenant(authContext(t, _userId), _createTenantReq)

	require.Error(t, err)
	assert.NoError(t, conn.ExpectationsWereMet())
}

func TestCreateTenantWithoutToken(t *testing.T) {
	_, err := New(nil, nil, nil, nil).
		CreateTenant(t.Context(), _createTenantReq)

	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestGetTenant(t *testing.T) {
	want := &tenantpb.Tenant{Id: _tenantId}

	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().Get(mock.Anything, _tenantId).Return(want, nil)

	resp, err := New(nil, nil, nil, repo).GetTenant(t.Context(),
		&tenantpb.GetTenantRequest{Id: _tenantId})

	require.NoError(t, err)
	assert.Equal(t, want, resp.GetTenant())
}

// missingTenant returns a service whose tenant repo has no rows.
func missingTenant(t *testing.T) *Service {
	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().Get(mock.Anything, _tenantId).Return(nil, pgx.ErrNoRows)

	return New(nil, nil, nil, repo)
}

func TestGetTenantNotFound(t *testing.T) {
	_, err := missingTenant(t).GetTenant(t.Context(),
		&tenantpb.GetTenantRequest{Id: _tenantId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestUpdateTenantNotFound(t *testing.T) {
	_, err := missingTenant(t).UpdateTenant(t.Context(),
		&tenantpb.UpdateTenantRequest{Id: _tenantId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteTenantNotFound(t *testing.T) {
	_, err := missingTenant(t).DeleteTenant(t.Context(),
		&tenantpb.DeleteTenantRequest{Id: _tenantId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestUpdateTenant(t *testing.T) {
	want := &tenantpb.Tenant{
		Id: _tenantId, Name: "acme", Plan: typepb.Plan_PLAN_PRO,
	}

	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().Get(mock.Anything, _tenantId).
		Return(&tenantpb.Tenant{Id: _tenantId}, nil)
	repo.EXPECT().Update(mock.Anything, want).Return(nil)

	_, err := New(nil, nil, nil, repo).UpdateTenant(t.Context(),
		&tenantpb.UpdateTenantRequest{
			Id: _tenantId, Name: "acme", Plan: typepb.Plan_PLAN_PRO,
		})

	require.NoError(t, err)
}

func TestDeleteTenant(t *testing.T) {
	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().Get(mock.Anything, _tenantId).
		Return(&tenantpb.Tenant{Id: _tenantId}, nil)
	repo.EXPECT().Delete(mock.Anything, _tenantId).Return(nil)

	_, err := New(nil, nil, nil, repo).DeleteTenant(t.Context(),
		&tenantpb.DeleteTenantRequest{Id: _tenantId})

	require.NoError(t, err)
}

func TestListTenant(t *testing.T) {
	req := &tenantpb.ListTenantRequest{Plan: typepb.Plan_PLAN_FREE}
	want := &tenantpb.ListTenantResponse{
		Tenants: []*tenantpb.Tenant{{Id: _tenantId}},
	}

	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().List(mock.Anything, req).Return(want, nil)

	resp, err := New(nil, nil, nil, repo).ListTenant(t.Context(), req)

	require.NoError(t, err)
	assert.Equal(t, want, resp)
}
