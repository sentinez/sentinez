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
	membermock "github.com/sentinez/controlplane/apps/tenant/v1/repos/members/mock"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	_memberId = "mem_1"
	_userId   = "user_1"
	_tenantId = "tnt_1"
)

func TestCreateMember(t *testing.T) {
	roles := []typepb.Role{typepb.Role_ROLE_MEMBER}
	want := &tenantpb.Member{
		UserId: _userId, TenantId: _tenantId, Roles: roles,
	}

	repo := membermock.NewMockIMember(t)
	// only this user in this tenant: other members must not block the invite
	repo.EXPECT().List(mock.Anything, &tenantpb.ListMemberRequest{
		TenantIds: []string{_tenantId}, UserId: _userId,
	}).Return(&tenantpb.ListMemberResponse{}, nil)
	repo.EXPECT().Create(mock.Anything, want).Return(want, nil)

	resp, err := New(nil, nil, repo, nil).CreateMember(t.Context(),
		&tenantpb.CreateMemberRequest{
			UserId: _userId, TenantId: _tenantId, Roles: roles,
		})

	require.NoError(t, err)
	assert.Equal(t, want, resp.GetMember())
}

func TestCreateMemberAlreadyExists(t *testing.T) {
	repo := membermock.NewMockIMember(t)
	repo.EXPECT().List(mock.Anything, mock.Anything).Return(
		&tenantpb.ListMemberResponse{
			Members: []*tenantpb.Member{{Id: _memberId, UserId: _userId}},
		}, nil)

	_, err := New(nil, nil, repo, nil).CreateMember(t.Context(),
		&tenantpb.CreateMemberRequest{UserId: _userId, TenantId: _tenantId})

	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestGetMember(t *testing.T) {
	want := &tenantpb.Member{Id: _memberId}

	repo := membermock.NewMockIMember(t)
	repo.EXPECT().Get(mock.Anything, _memberId).Return(want, nil)

	resp, err := New(nil, nil, repo, nil).GetMember(t.Context(),
		&tenantpb.GetMemberRequest{Id: _memberId})

	require.NoError(t, err)
	assert.Equal(t, want, resp.GetMember())
}

// missingMember returns a service whose member repo has no rows.
func missingMember(t *testing.T) *Service {
	repo := membermock.NewMockIMember(t)
	repo.EXPECT().Get(mock.Anything, _memberId).Return(nil, pgx.ErrNoRows)

	return New(nil, nil, repo, nil)
}

func TestGetMemberNotFound(t *testing.T) {
	_, err := missingMember(t).GetMember(t.Context(),
		&tenantpb.GetMemberRequest{Id: _memberId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestUpdateMemberNotFound(t *testing.T) {
	_, err := missingMember(t).UpdateMember(t.Context(),
		&tenantpb.UpdateMemberRequest{Id: _memberId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteMemberNotFound(t *testing.T) {
	_, err := missingMember(t).DeleteMember(t.Context(),
		&tenantpb.DeleteMemberRequest{Id: _memberId})

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestUpdateMember(t *testing.T) {
	roles := []typepb.Role{typepb.Role_ROLE_ADMIN}

	repo := membermock.NewMockIMember(t)
	repo.EXPECT().Get(mock.Anything, _memberId).
		Return(&tenantpb.Member{Id: _memberId}, nil)
	repo.EXPECT().Update(mock.Anything,
		&tenantpb.Member{Id: _memberId, Roles: roles}).Return(nil)

	_, err := New(nil, nil, repo, nil).UpdateMember(t.Context(),
		&tenantpb.UpdateMemberRequest{Id: _memberId, Roles: roles})

	require.NoError(t, err)
}

func TestDeleteMember(t *testing.T) {
	repo := membermock.NewMockIMember(t)
	repo.EXPECT().Get(mock.Anything, _memberId).
		Return(&tenantpb.Member{Id: _memberId}, nil)
	repo.EXPECT().Delete(mock.Anything, _memberId).Return(nil)

	_, err := New(nil, nil, repo, nil).DeleteMember(t.Context(),
		&tenantpb.DeleteMemberRequest{Id: _memberId})

	require.NoError(t, err)
}

func TestListMember(t *testing.T) {
	req := &tenantpb.ListMemberRequest{TenantIds: []string{_tenantId}}
	want := &tenantpb.ListMemberResponse{
		Members: []*tenantpb.Member{{Id: _memberId}},
	}

	repo := membermock.NewMockIMember(t)
	repo.EXPECT().List(mock.Anything, req).Return(want, nil)

	resp, err := New(nil, nil, repo, nil).ListMember(t.Context(), req)

	require.NoError(t, err)
	assert.Equal(t, want, resp)
}
