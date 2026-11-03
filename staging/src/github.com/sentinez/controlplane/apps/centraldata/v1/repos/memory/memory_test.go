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

package memrepo

import (
	"testing"

	"github.com/sentinez/controlplane/pkg/schema"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	"github.com/sentinez/shared/szerror"
	"github.com/sentinez/shared/sztable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMemory(t *testing.T) *Memory {
	t.Helper()

	m, err := New()
	require.NoError(t, err)

	return m
}

func TestNewHasEveryTable(t *testing.T) {
	m := newMemory(t)

	for _, table := range schema.All() {
		t.Run(table.Table.String(), func(t *testing.T) {
			_, err := List[struct{}](m, table.Table)

			require.NoError(t, err)
		})
	}
}

func TestUpsertGet(t *testing.T) {
	m := newMemory(t)

	require.NoError(t, m.Upsert(sztable.Tenants,
		&tenantpb.Tenant{Id: "tn-1", Name: "a"}))
	require.NoError(t, m.Upsert(sztable.Tenants,
		&tenantpb.Tenant{Id: "tn-1", Name: "b"}))

	got, err := Get[tenantpb.Tenant](m, sztable.Tenants, "tn-1")

	require.NoError(t, err)
	assert.Equal(t, "b", got.GetName())
}

func TestGetNotFound(t *testing.T) {
	_, err := Get[tenantpb.Tenant](newMemory(t), sztable.Tenants, "x")

	require.ErrorIs(t, err, szerror.ErrNotFound)
}

func TestGetWrongType(t *testing.T) {
	m := newMemory(t)
	require.NoError(t, m.Upsert(sztable.Tenants, &tenantpb.Tenant{Id: "1"}))

	_, err := Get[tenantpb.Member](m, sztable.Tenants, "1")

	require.ErrorContains(t, err, "not *tenantpb.Member")
}

func TestUpsertAllOrNone(t *testing.T) {
	m := newMemory(t)

	err := m.Upsert(sztable.Tenants,
		&tenantpb.Tenant{Id: "1"}, &tenantpb.Tenant{})

	require.Error(t, err)
	rows, err := List[tenantpb.Tenant](m, sztable.Tenants)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestListDelete(t *testing.T) {
	m := newMemory(t)
	require.NoError(t, m.Upsert(sztable.Tenants,
		&tenantpb.Tenant{Id: "2"}, &tenantpb.Tenant{Id: "1"}))

	require.NoError(t, m.Delete(sztable.Tenants, "2"))
	require.ErrorIs(t, m.Delete(sztable.Tenants, "2"), szerror.ErrNotFound)

	rows, err := List[tenantpb.Tenant](m, sztable.Tenants)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "1", rows[0].GetId())
}
