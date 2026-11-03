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
	"fmt"

	"github.com/hashicorp/go-memdb"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/shared/szerror"
	"github.com/sentinez/shared/szstorage/dbx"
	"github.com/sentinez/shared/sztable"
)

// Memory the in-memory store, safe for concurrent use: go-memdb serves
// reads from immutable snapshots and serializes the writes.
type Memory struct {
	db *memdb.MemDB
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

	return &Memory{db: db}, nil
}

// Upsert inserts the rows into table, replacing those of the same id,
// all or none. A row is a pointer to the model of the table, ex:
// *tenantpb.Tenant for sztable.Tenants.
func (m *Memory) Upsert(table sztable.Table, rows ...any) error {
	if len(rows) == 0 {
		return nil
	}

	name := table.String()
	txn := m.db.Txn(true)
	defer txn.Abort()

	for _, row := range rows {
		if err := txn.Insert(name, row); err != nil {
			return fmt.Errorf("upsert %s: %w", table, err)
		}
	}

	txn.Commit()

	return nil
}

// Delete removes the row of id from table.
func (m *Memory) Delete(table sztable.Table, id string) error {
	txn := m.db.Txn(true)
	defer txn.Abort()

	row, err := first(txn, table, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := txn.Delete(table.String(), row); err != nil {
		return fmt.Errorf("delete %s: %w", table, err)
	}

	txn.Commit()

	return nil
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
