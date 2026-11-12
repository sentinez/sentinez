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

// Package memrepo keeps the controlplane data of centraldata in memory, in
// a go-memdb table per controlplane table, built from schema.All.
package memrepo

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/hashicorp/go-memdb"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/sztable"
)

// Memory the in-memory store, safe for concurrent use: go-memdb serves
// reads from immutable snapshots and serializes the writes.
//
// Every write bumps a version, so a reader of Snapshot knows which
// changes the rows it holds already include.
type Memory struct {
	db *memdb.MemDB

	// mu orders a write with its version bump, and Snapshot with both.
	mu      sync.Mutex
	version uint64
	epoch   string
}

// New creates the Memory with a table per schema of schema.All.
func New() (*Memory, error) {
	dbSchema, err := dbx.MemDBSchema(schema.All()...)
	if err != nil {
		return nil, err
	}

	db, err := memdb.NewMemDB(dbSchema)
	if err != nil {
		return nil, fmt.Errorf("new memdb: %w", err)
	}

	return &Memory{db: db, epoch: newEpoch()}, nil
}

// Epoch identifies the content lineage of the versions: Reset changes
// it, and the versions of two epochs are not comparable.
func (m *Memory) Epoch() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.epoch
}

// Version the number of writes since the last Reset.
func (m *Memory) Version() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.version
}

// Reset starts a new epoch at version 0, the rows are kept: the caller
// reloads them.
func (m *Memory) Reset() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.epoch, m.version = newEpoch(), 0

	return m.epoch
}

// Upsert inserts the rows into table, replacing those of the same id,
// all or none. A row is a pointer to the model of the table, ex:
// *tenantpb.Tenant for sztable.Tenants.
func (m *Memory) Upsert(table sztable.Table, rows ...any) error {
	_, err := m.UpsertAt(table, rows...)

	return err
}

// UpsertAt is Upsert returning the version of the write; the current
// one when there are no rows.
func (m *Memory) UpsertAt(
	table sztable.Table, rows ...any) (uint64, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(rows) == 0 {
		return m.version, nil
	}

	name := table.String()
	txn := m.db.Txn(true)
	defer txn.Abort()

	for _, row := range rows {
		if err := txn.Insert(name, row); err != nil {
			return 0, fmt.Errorf("upsert %s: %w", table, err)
		}
	}

	txn.Commit()
	m.version++

	return m.version, nil
}

// Delete removes the row of id from table.
func (m *Memory) Delete(table sztable.Table, id string) error {
	_, err := m.DeleteAt(table, id)

	return err
}

// DeleteAt is Delete returning the version of the write.
func (m *Memory) DeleteAt(table sztable.Table, id string) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	txn := m.db.Txn(true)
	defer txn.Abort()

	row, err := first(txn, table, id)
	if err != nil {
		return 0, fmt.Errorf("delete: %w", err)
	}

	if err := txn.Delete(table.String(), row); err != nil {
		return 0, fmt.Errorf("delete %s: %w", table, err)
	}

	txn.Commit()
	m.version++

	return m.version, nil
}

// Snapshot calls fn for every row of tables, ordered by table then id,
// from one consistent view, and returns the version and epoch of that
// view. A change published with a greater version in the same epoch is
// not in the rows fn saw.
func (m *Memory) Snapshot(tables []sztable.Table,
	fn func(table sztable.Table, row any) error) (uint64, string, error) {

	m.mu.Lock()
	txn := m.db.Txn(false)
	version, epoch := m.version, m.epoch
	m.mu.Unlock()

	for _, table := range tables {
		it, err := txn.Get(table.String(), dbx.MemIdIndex)
		if err != nil {
			return 0, "", fmt.Errorf("snapshot %s: %w", table, err)
		}

		for raw := it.Next(); raw != nil; raw = it.Next() {
			if err := fn(table, raw); err != nil {
				return 0, "", err
			}
		}
	}

	return version, epoch, nil
}

// Get the row of id in table, as a *T.
func Get[T any](m *Memory, table sztable.Table, id string) (*T, error) {
	raw, err := first(m.db.Txn(false), table, id)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return as[T](table, raw)
}

// List the rows of table ordered by id, as *T.
func List[T any](m *Memory, table sztable.Table) ([]*T, error) {
	it, err := m.db.Txn(false).Get(table.String(), dbx.MemIdIndex)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", table, err)
	}

	var rows []*T
	for raw := it.Next(); raw != nil; raw = it.Next() {
		row, err := as[T](table, raw)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// first the row of id in table, szerror.ErrNotFound when there is none.
func first(txn *memdb.Txn, table sztable.Table, id string) (any, error) {
	row, err := txn.First(table.String(), dbx.MemIdIndex, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", table, err)
	}

	if row == nil {
		return nil, fmt.Errorf("%s %s: %w", table, id, szerror.ErrNotFound)
	}

	return row, nil
}

// as asserts raw, a row of table, is a *T.
func as[T any](table sztable.Table, raw any) (*T, error) {
	row, ok := raw.(*T)
	if !ok {
		return nil, fmt.Errorf("%s holds %T, not %T", table, raw, row)
	}

	return row, nil
}

// newEpoch a random identifier of a content lineage.
func newEpoch() string {
	var b [8]byte
	_, _ = rand.Read(b[:])

	return hex.EncodeToString(b[:])
}
