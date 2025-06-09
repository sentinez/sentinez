// Copyright 2025 Duc-Hung Ho.
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

package pg

import (
	"github.com/jackc/pgx/v5"
	"github.com/sentinez/sentinez/pkg/infra/driver/db"
)

// make sure dbRows implement db.Rows
var _ db.Rows[int] = (*dbRows[int])(nil)

type dbRows[T any] struct {
	rows pgx.Rows
}

// CollectOne collects a single row from the database and maps it
// to the specified type.
func (r *dbRows[T]) CollectOne() (T, error) {
	return pgx.CollectOneRow(r.rows, pgx.RowToStructByName[T])
}

// CollectAll collects all rows from the database and maps them
// to a slice of the specified type.
func (r *dbRows[T]) CollectAll() ([]T, error) {
	return pgx.CollectRows(r.rows, pgx.RowToStructByName[T])
}
