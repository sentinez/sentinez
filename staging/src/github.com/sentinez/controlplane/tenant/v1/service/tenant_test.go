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
	"testing"

	"github.com/jackc/pgx/v5"
	tenantmock "github.com/sentinez/controlplane/tenant/v1/repos/tenants/mock"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateTenant(t *testing.T) {
	want := &tenantpb.Tenant{
		Name:        "acme",
		Description: "acme team",
		Plan:        typepb.Plan_PLAN_FREE,
	}

	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().Create(mock.Anything, want).Return(want, nil)

	resp, err := New(nil, nil, repo).CreateTenant(t.Context(),
		&tenantpb.CreateTenantRequest{
			Name:        "acme",
			Description: "acme team",
			Plan:        typepb.Plan_PLAN_FREE,
		})

	require.NoError(t, err)
	assert.Equal(t, want, resp.GetTenant())
}

func TestGetTenant(t *testing.T) {
	want := &tenantpb.Tenant{Id: _tenantId}

	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().Get(mock.Anything, _tenantId).Return(want, nil)

	resp, err := New(nil, nil, repo).GetTenant(t.Context(),
		&tenantpb.GetTenantRequest{Id: _tenantId})

	require.NoError(t, err)
	assert.Equal(t, want, resp.GetTenant())
}

// missingTenant returns a service whose tenant repo has no rows.
func missingTenant(t *testing.T) *Service {
	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().Get(mock.Anything, _tenantId).Return(nil, pgx.ErrNoRows)

	return New(nil, nil, repo)
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

	_, err := New(nil, nil, repo).UpdateTenant(t.Context(),
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

	_, err := New(nil, nil, repo).DeleteTenant(t.Context(),
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

	resp, err := New(nil, nil, repo).ListTenant(t.Context(), req)

	require.NoError(t, err)
	assert.Equal(t, want, resp)
}
