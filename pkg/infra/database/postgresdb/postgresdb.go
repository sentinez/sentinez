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

package postgresdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinez/sentinez/pkg/infra/database"
	pgopt "github.com/sentinez/sentinez/pkg/infra/options/postgres"
	"github.com/sentinez/sentinez/pkg/std/table"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var _ database.Database[struct{}] = (*postgres[struct{}])(nil)

func New[T any](pool *pgxpool.Pool, tableName string,
	opts ...database.Option) (database.Database[T], error) {

	if table.IsValidTableName(tableName) == false {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	table := database.Table{}
	for _, opt := range opts {
		opt(&table)
	}

	switch table.StorageOption {
	case database.StorageKV:
		if err := pgopt.TableKV(
			pool, strings.ReplaceAll(tableName, ".", "_"),
		); err != nil {
			zlog.Debug("[postgresdb] create err: ", err)

			return nil, fmt.Errorf("failed to create table %s", tableName)
		}
	}

	for _, ref := range table.References {
		err := pgopt.Reference(pool,
			tableName, ref.FromField, ref.ToTable, ref.ToField)
		if err != nil {
			zlog.Debug("[postgresdb] reference err: ", err)

			return nil, fmt.Errorf("failed to create reference %s.%s -> %s.%s",
				tableName, ref.FromField, ref.ToTable, ref.ToField)
		}
	}

	return &postgres[T]{
		pool: pool,
	}, nil
}

type postgres[T any] struct {
	pool *pgxpool.Pool
}

// BeginTx implements database.Database.
func (p *postgres[T]) BeginTx(
	ctx context.Context) (database.Transaction[T], error) {

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &PostgresTx[T]{tx: tx}, nil
}

// Collect implements database.Database.
func (p *postgres[T]) CollectRows(ctx context.Context,
	builder database.SQLBuilder,
	_ func(database.Rows) ([]T, error)) ([]T, error) {

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[T])
}

// Collect implements database.Database.
func (p *postgres[T]) Collect(ctx context.Context,
	builder database.SQLBuilder, _ func(database.Row) (*T, error)) (*T, error) {

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Exec implements database.Database.
func (p *postgres[T]) Exec(ctx context.Context,
	builder database.SQLBuilder) (database.ExecResult, error) {

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	result, err := p.pool.Exec(ctx, sqlStr, args...)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return result, nil
}

// WithTx implements database.Database.
func (p *postgres[T]) WithTx(ctx context.Context,
	fn func(database.Transaction[T]) error) error {

	tx, err := p.BeginTx(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
