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

	"github.com/sentinez/sentinez/pkg/infra/database/query"
	"google.golang.org/protobuf/proto"
)

const Data = "data"

// Repository provides the interface for the database.
type Repository[T proto.Message, ID comparable] interface {
	Create(ctx context.Context, entity T) (T, error)
	Get(ctx context.Context, id ID) (T, error)
	Update(ctx context.Context, entity T) (T, error)
	Delete(ctx context.Context, id ID) error
}

type Executor[T proto.Message] interface {
	Exec(ctx context.Context, builder query.Query) (ExecResult, error)
	Query(ctx context.Context, builder query.Query, dest ...any) error
	CollectRows(ctx context.Context, builder query.Query,
		scan func(Rows) ([]T, error)) ([]T, error)
	CollectOneRow(ctx context.Context, builder query.Query,
		scan func(Row) (T, error)) (T, error)
}

type Database[T proto.Message] interface {
	Executor[T]
	BeginTx(ctx context.Context) (Transaction[T], error)
	WithTx(ctx context.Context, fn func(Transaction[T]) error) error
}

type Transaction[T proto.Message] interface {
	Exec(ctx context.Context, builder query.Query) (ExecResult, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	CollectRows(ctx context.Context, builder query.Query,
		scan func(Rows) ([]T, error)) ([]T, error)
}

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	// Close() error
}

type ExecResult interface {
	RowsAffected() int64
}

type Table struct {
	Index []string
}

type Option func(*Table)
