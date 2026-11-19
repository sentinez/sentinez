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

package apiserver

import (
	"context"
	"errors"
	"os"
	"testing"

	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTenant records the imported resource.
type fakeTenant struct {
	err      error
	tenant   *tenantpb.Tenant
	resource *edgepb.Resource
}

func (f *fakeTenant) ImportResource(_ context.Context, tenant *tenantpb.Tenant,
	srv *edgepb.Resource) error {
	f.tenant, f.resource = tenant, srv
	return f.err
}

type fakeSecurity struct {
	owner *typepb.Owner
	name  string
	sec   *edgepb.Security
}

func (f *fakeSecurity) ImportSecurity(_ context.Context,
	owner *typepb.Owner, name string, sec *edgepb.Security) error {
	f.owner, f.name, f.sec = owner, name, sec
	return nil
}

// readProxy returns the proxy.yaml the edge ships with.
func readProxy(t *testing.T) []byte {
	t.Helper()

	data, err := os.ReadFile("../../../../cmd/szedge/v1/proxy.yaml")
	require.NoError(t, err)
	return data
}

func TestStoreProxy(t *testing.T) {
	tenant, security := &fakeTenant{}, &fakeSecurity{}

	err := storeProxy(t.Context(), readProxy(t), tenant, security)

	require.NoError(t, err)
	assert.Equal(t, "tnt_default_admin", tenant.tenant.GetId())
	assert.Equal(t, "admin", tenant.tenant.GetName())
	assert.Equal(t, typepb.Plan_PLAN_PRO, tenant.tenant.GetPlan())
	assert.Equal(t, "badcheese", tenant.resource.GetName())
	assert.Equal(t, "badcheese.is.s6z.io.vn",
		tenant.resource.GetDomain())
	assert.Equal(t, "rsrc_default", tenant.resource.GetId())
	assert.Equal(t, "badcheese", security.name)
	assert.Equal(t, &typepb.Owner{
		TenantId: "tnt_default_admin", ResourceId: "rsrc_default",
	}, security.owner)
	require.NotEmpty(t, security.sec.GetSecrules())
	assert.NotNil(t, security.sec.GetSecrules()[0].GetIngress())
	assert.NotNil(t, security.sec.GetLimiters()[0].GetIngress())
}

func TestStoreProxyDefaults(t *testing.T) {
	tenant := &fakeTenant{}

	data := []byte("setting: {ingress: {resource: {name: a}}}")
	err := storeProxy(t.Context(), data, tenant, &fakeSecurity{})

	require.NoError(t, err)
	assert.Equal(t, &tenantpb.Tenant{
		Id: defaults.TenantId, Name: defaults.TenantId,
	}, tenant.tenant)
	assert.Equal(t, defaults.ResourceId, tenant.resource.GetId())
}

func TestStoreProxyErrors(t *testing.T) {
	tests := []struct {
		name       string
		giveData   []byte
		giveTenant *fakeTenant
	}{
		{"no setting", []byte("other: 1"), &fakeTenant{}},
		{"bad yaml", []byte("setting: ["), &fakeTenant{}},
		{"no resource", []byte("setting: {metadata: {}}"), &fakeTenant{}},
		{"tenant fails", []byte("setting: {ingress: {resource: {name: a}}}"),
			&fakeTenant{err: errors.New("db down")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			security := &fakeSecurity{}

			err := storeProxy(t.Context(), tt.giveData, tt.giveTenant,
				security)

			require.Error(t, err)
			assert.Nil(t, security.owner)
		})
	}
}
