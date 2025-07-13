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
package database

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
)

type StorageOption int

const (
	StorageKV StorageOption = iota
)

// Repository provides the interface for the database.
type Repository[T any, ID comparable] interface {
	Create(ctx context.Context, entity T) (T, error)
	Get(ctx context.Context, id ID) (T, error)
	GetMany(ctx context.Context, page *common.Pages) ([]T, error)
	Update(ctx context.Context, entity T) (T, error)
	Delete(ctx context.Context, id ID) error
	Exists(ctx context.Context, id ID) (bool, error)
	Count(ctx context.Context) (int64, error)
}

type SQLBuilder interface {
	ToSql() (string, []any, error)
}

type Database[T any] interface {
	Exec(ctx context.Context, builder SQLBuilder) (ExecResult, error)

	CollectRows(ctx context.Context,
		builder SQLBuilder, scan func(Rows) ([]T, error)) ([]T, error)

	Collect(ctx context.Context,
		builder SQLBuilder, scan func(Row) (*T, error)) (*T, error)

	BeginTx(ctx context.Context) (Transaction[T], error)
	WithTx(ctx context.Context, fn func(Transaction[T]) error) error
}

type Transaction[T any] interface {
	Exec(ctx context.Context, builder SQLBuilder) (ExecResult, error)

	CollectRows(ctx context.Context,
		builder SQLBuilder, scan func(Rows) ([]T, error)) ([]T, error)

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
	Close() error
}

type ExecResult interface {
	RowsAffected() int64
}

type Reference struct {
	FromField string
	ToTable   string
	ToField   string
}

type Table struct {
	References    map[string]Reference // table name -> references
	StorageOption StorageOption
}

type Option func(*Table)
