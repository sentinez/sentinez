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
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/sztable"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
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

func TestWritesBumpVersion(t *testing.T) {
	m := newMemory(t)

	up, err := m.UpsertAt(sztable.Tenants, &tenantpb.Tenant{Id: "tn-1"})
	require.NoError(t, err)
	del, err := m.DeleteAt(sztable.Tenants, "tn-1")
	require.NoError(t, err)

	assert.Equal(t, uint64(1), up)
	assert.Equal(t, uint64(2), del)
	assert.Equal(t, uint64(2), m.Version())
}

func TestFailedDeleteKeepsVersion(t *testing.T) {
	m := newMemory(t)

	_, err := m.DeleteAt(sztable.Tenants, "x")

	require.ErrorIs(t, err, szerror.ErrNotFound)
	assert.Zero(t, m.Version())
}

func TestResetChangesEpoch(t *testing.T) {
	m := newMemory(t)
	before := m.Epoch()
	require.NoError(t, m.Upsert(sztable.Tenants, &tenantpb.Tenant{Id: "a"}))

	after := m.Reset()

	assert.NotEqual(t, before, after)
	assert.Equal(t, after, m.Epoch())
	assert.Zero(t, m.Version())
}

func TestSnapshot(t *testing.T) {
	m := newMemory(t)
	require.NoError(t, m.Upsert(sztable.Tenants,
		&tenantpb.Tenant{Id: "b"}, &tenantpb.Tenant{Id: "a"}))

	var ids []string
	version, epoch, err := m.Snapshot([]sztable.Table{sztable.Tenants},
		func(_ sztable.Table, row any) error {
			ids = append(ids, row.(*tenantpb.Tenant).GetId())

			return nil
		})

	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, ids)
	assert.Equal(t, uint64(1), version)
	assert.Equal(t, m.Epoch(), epoch)
}
