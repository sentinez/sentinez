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
	"errors"
	"testing"

	resourcemock "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources/mock"
	tenantmock "github.com/sentinez/controlplane/apps/tenant/v1/repos/tenants/mock"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	_importServer = &edgepb.Server{Name: "badcheese", ResourceId: "rsrc_shop"}
	_importTenant = &tenantpb.Tenant{
		Id: _tenantId, Name: "admin", Plan: typepb.Plan_PLAN_PRO,
	}
)

// importTenants expects _importTenant to be upserted.
func importTenants(t *testing.T) *tenantmock.MockITenant {
	t.Helper()

	repo := tenantmock.NewMockITenant(t)
	repo.EXPECT().WithTX(mock.Anything).Return(repo)
	repo.EXPECT().Upsert(mock.Anything, _importTenant).Return(nil)

	return repo
}

func TestImportServer(t *testing.T) {
	tx, conn := mockTx(t)
	conn.ExpectCommit()

	want := &tenantpb.Resource{
		Id: "rsrc_shop", TenantId: _tenantId,
		ResourceName:    "badcheese",
		ResourceDomain:  "badcheese.s6z.io.vn",
		ResourceSetting: &tenantpb.ResourceSetting{Server: _importServer},
		Status:          typepb.Status_STATUS_ACTIVE,
	}
	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().WithTX(mock.Anything).Return(resources)
	resources.EXPECT().Upsert(mock.Anything, want).Return(nil)

	err := New(tx, resources, nil, importTenants(t)).ImportServer(
		t.Context(), _importTenant, "s6z.io.vn", _importServer)

	require.NoError(t, err)
	assert.NoError(t, conn.ExpectationsWereMet())
}

func TestImportServerRollsBackWhenResourceFails(t *testing.T) {
	tx, conn := mockTx(t)
	conn.ExpectRollback()

	resources := resourcemock.NewMockIResource(t)
	resources.EXPECT().WithTX(mock.Anything).Return(resources)
	resources.EXPECT().Upsert(mock.Anything, mock.Anything).
		Return(errors.New("insert resource"))

	err := New(tx, resources, nil, importTenants(t)).ImportServer(
		t.Context(), _importTenant, "s6z.io.vn", _importServer)

	require.Error(t, err)
	assert.NoError(t, conn.ExpectationsWereMet())
}

func TestResourceDomain(t *testing.T) {
	tests := []struct {
		name         string
		giveHostname string
		want         string
	}{
		{"hostname", "s6z.io.vn", "shop.s6z.io.vn"},
		{"no hostname", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resourceDomain("shop", tt.giveHostname))
		})
	}
}
