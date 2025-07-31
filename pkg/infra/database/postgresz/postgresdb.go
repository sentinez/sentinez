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
	"reflect"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/infra/database/query"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/table"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/encoding/protojson"
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

	err := syncProtoToPostgresJSONB(
		context.Background(), pool, tableName, tb.Index)
	if err != nil {
		return nil, fmt.Errorf("failed to sync proto to postgres: %w", err)
	}

	return &postgres[T]{pool: pool, tableName: tableName}, nil
}

type postgres[T proto.Message] struct {
	pool      *pgxpool.Pool
	tableName string
}

func (p *postgres[T]) Set(ctx context.Context, id string, entity T) error {
	data, err := protojson.Marshal(entity)
	if err != nil {
		return nil
	}

	builder := sq.Insert(p.tableName).
		Columns(database.ID, database.Data).
		Values(id, string(data)).
		Suffix(fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET data = EXCLUDED.%s",
			database.ID, database.Data,
		))

	_, err = p.Exec(ctx, builder)
	return err
}

func (p *postgres[T]) Get(ctx context.Context, id string) (T, error) {
	var (
		empty T
		data  string
	)

	result := reflect.New(
		reflect.TypeOf((*T)(nil)).Elem().Elem()).Interface().(proto.Message)

	builder := sq.Select(database.Data).From(p.tableName).Where(sq.Eq{
		database.ID: id,
	})

	if err := p.Query(ctx, builder, &data); err != nil {
		return empty, err
	}

	if err := protojson.Unmarshal([]byte(data), result); err != nil {
		return empty, err
	}

	return result.(T), nil
}

func (p *postgres[T]) Delete(ctx context.Context, id string) error {
	builder := sq.Delete(p.tableName).Where(sq.Eq{
		database.ID: id,
	})

	_, err := p.Exec(ctx, builder)
	return err
}

// BeginTx implements database.Database.
func (p *postgres[T]) BeginTx(
	ctx context.Context) (database.Transaction[T], error) {

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &postgresTx[T]{tx: tx, tableName: p.tableName}, nil
}

// CollectRows implements database.Database.
func (p *postgres[T]) CollectRows(ctx context.Context,
	builder query.Query,
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

	return nil, errors.F("[CollectRows] missing scans function")
}

// Collect implements database.Database.
func (p *postgres[T]) CollectOneRow(ctx context.Context,
	builder query.Query, scan func(database.Row) (T, error)) (T, error) {

	var empty T

	query, args, err := builder.ToSql()
	if err != nil {
		return empty, err
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	row := p.pool.QueryRow(ctx, query, args...)

	if scan != nil {
		return scan(row)
	}

	return empty, errors.F("[CollectOneRow] missing scans function")
}

// Exec implements database.Database.
func (p *postgres[T]) Exec(ctx context.Context,
	builder query.Query) (database.ExecResult, error) {

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

func (p *postgres[T]) Query(ctx context.Context,
	builder query.Query, dest ...any) error {

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	result, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return err
	}

	for result.Next() {
		if err := result.Scan(dest...); err != nil {
			return err
		}
	}

	return nil
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
