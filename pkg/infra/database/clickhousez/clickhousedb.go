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

package clickhousez

import (
	"context"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/infra/database/query"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"google.golang.org/protobuf/proto"
)

var _ database.Database[*common.Empty] = (*clickHouse[*common.Empty])(nil)

func New[T proto.Message](conn clickhouse.Conn) database.Database[T] {
	return &clickHouse[T]{
		conn: conn,
	}
}

type clickHouse[T proto.Message] struct {
	conn clickhouse.Conn
}

// BeginTx implements database.Database.
func (c *clickHouse[T]) BeginTx(
	_ context.Context) (database.Transaction[T], error) {
	return nil, errors.F("[sentinez] clickhouse transaction not supported")
}

// CollectRows implements database.Database.
func (c *clickHouse[T]) CollectRows(ctx context.Context,
	builder query.Query,
	scan func(database.Rows) ([]T, error)) ([]T, error) {

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scan(rows)
}

func (c *clickHouse[T]) CollectOneRow(ctx context.Context,
	builder query.Query,
	scan func(database.Row) (T, error)) (T, error) {

	var empty T

	query, args, err := builder.ToSql()
	if err != nil {
		return empty, err
	}

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return empty, err
	}
	defer func() { _ = rows.Close() }()

	return scan(rows)
}

// Exec implements database.Database.
func (c *clickHouse[T]) Exec(ctx context.Context,
	builder query.Query) (database.ExecResult, error) {

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	if err := c.conn.Exec(ctx, query, args...); err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *clickHouse[T]) Query(ctx context.Context,
	builder query.Query, dest ...any) error {

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	result, err := c.conn.Query(ctx, query, args...)
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
func (c *clickHouse[T]) WithTx(_ context.Context,
	_ func(database.Transaction[T]) error) error {

	return errors.F("[sentinez] clickhouse transaction not supported")
}

func (c *clickHouse[T]) SelectBuilder() squirrel.SelectBuilder {
	// Replace "table_name" with actual table name
	return squirrel.Select("data").From("table_name")
}
