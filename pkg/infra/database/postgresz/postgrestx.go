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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/infra/database/query"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var _ database.Transaction[*common.Empty] = (*postgresTx[*common.Empty])(nil)

type postgresTx[T proto.Message] struct {
	tx        pgx.Tx
	tableName string
}

func (p *postgresTx[T]) Set(ctx context.Context, id string, entity T) error {
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

func (p *postgresTx[T]) Get(ctx context.Context, id string) (T, error) {
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

func (p *postgresTx[T]) Delete(ctx context.Context, id string) error {
	builder := sq.Delete(p.tableName).Where(sq.Eq{
		database.ID: id,
	})

	_, err := p.Exec(ctx, builder)
	return err
}

// CollectRows implements database.Transaction.
func (p *postgresTx[T]) CollectRows(ctx context.Context,
	builder query.Query, _ func(database.Rows) ([]T, error)) ([]T, error) {

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.tx.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[T])
}

// Commit implements database.Transaction.
func (p *postgresTx[T]) Commit(ctx context.Context) error {
	if err := p.tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// Exec implements database.Transaction.
func (p *postgresTx[T]) Exec(ctx context.Context,
	builder query.Query) (database.ExecResult, error) {

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	tag, err := p.tx.Exec(ctx, sqlStr, args...)
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	return tag, nil
}

func (p *postgresTx[T]) Query(ctx context.Context,
	builder query.Query, dest ...any) error {

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	result, err := p.tx.Query(ctx, query, args...)
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

// Rollback implements database.Transaction.
func (p *postgresTx[T]) Rollback(ctx context.Context) error {
	if err := p.tx.Rollback(ctx); err != nil {
		return err
	}

	return nil
}

func (p *postgresTx[T]) CollectOneRow(ctx context.Context,
	builder query.Query, scan func(database.Row) (T, error)) (T, error) {

	var empty T

	query, args, err := builder.ToSql()
	if err != nil {
		return empty, err
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	row := p.tx.QueryRow(ctx, query, args...)

	if scan != nil {
		return scan(row)
	}

	return empty, errors.F("[CollectOneRow] missing scans function")
}
