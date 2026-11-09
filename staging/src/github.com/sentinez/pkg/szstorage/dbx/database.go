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

// Package database provides the database interface.
package dbx

import (
	"context"

	"github.com/sentinez/pkg/szstorage/dbx/query"
)

type ScanOneFn[T any] func(Row) (*T, error)
type ScanFn[T any] func(Rows) ([]*T, error)

// nolint:lll
type Executor[T any] interface {
	Exec(ctx context.Context, builder query.Query) (ExecResult, error)
	Query(ctx context.Context, builder query.Query, dest ...any) error
	CollectRows(ctx context.Context, builder query.Query, scan ScanFn[T]) ([]*T, error)
	CollectOneRow(ctx context.Context, builder query.Query, scan ScanOneFn[T]) (*T, error)
}

// nolint:lll
type Database[T any] interface {
	Executor[T]
	Insert(ctx context.Context, builder query.Query) (string, error)
	Select(ctx context.Context, builder query.Query, scan ScanOneFn[T]) (*T, error)
	Delete(ctx context.Context, id string) error

	Total(ctx context.Context) (int64, error)

	Table() string
	// CreatedAt names the creation time column of the table.
	CreatedAt() string
	// UpdatedAt names the last update time column of the table.
	UpdatedAt() string
}

type TxSession interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

type ExecResult interface {
	RowsAffected() int64
}
