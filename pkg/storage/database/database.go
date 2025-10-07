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

	"github.com/sentinez/sentinez/pkg/storage/database/query"
	"google.golang.org/protobuf/proto"
)

const (
	SchemalessFieldData      = "data"
	SchemalessFieldID        = "id"
	SchemalessFieldCreatedAt = "created_at"
	SchemalessFieldUpdatedAt = "updated_at"
	SchemalessF              = `
	CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		data JSONB NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'UTC'),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'UTC')
	);
`
)

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
	Set(ctx context.Context, id string, entity T) error
	Get(ctx context.Context, id string) (T, error)
	Delete(ctx context.Context, id string) error
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
	// Close() error
}

type ExecResult interface {
	RowsAffected() int64
}

type Table struct {
	Index []string
}

type Option func(*Table)
