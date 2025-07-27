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

package postgresz

import (
	"context"
	"fmt"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/std/table"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

var _ database.Database[*common.Empty] = (*postgres[*common.Empty])(nil)

//nolint:funlen
func New[T proto.Message](pool *pgxpool.Pool, tableName string,
	opts ...database.Option) (database.Database[T], error) {

	if table.IsValidTableName(tableName) == false {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	tb := database.Table{}
	for _, opt := range opts {
		opt(&tb)
	}

	err := syncProtoToPostgresJSONB(context.Background(), pool,
		tableName,
		tb.Index,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to sync proto to postgres: %w", err)
	}

	return &postgres[T]{pool: pool, tableName: tableName}, nil
}

type postgres[T proto.Message] struct {
	pool      *pgxpool.Pool
	tableName string
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

// CollectRows implements database.Database.
func (p *postgres[T]) CollectRows(ctx context.Context,
	builder database.SQLBuilder,
	fn func(database.Rows) ([]T, error)) ([]T, error) {

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if fn != nil {
		return fn(rows)
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[T])
}

// Collect implements database.Database.
func (p *postgres[T]) Collect(ctx context.Context,
	builder database.SQLBuilder, _ func(database.Row) (*T, error)) (*T, error) {

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	rows, err := p.pool.Query(ctx, query, args...)
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

	query, args, err := builder.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	result, err := p.pool.Exec(ctx, query, args...)
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

func (p *postgres[T]) Total(ctx context.Context) (int64, error) {

	builder := sq.Select("COUNT(*) AS count").From(p.tableName)
	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	row := p.pool.QueryRow(ctx, query, args...)
	var count int64
	err = row.Scan(&count)
	return count, err
}
