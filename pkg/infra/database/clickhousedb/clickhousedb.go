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

package clickhousedb

import (
	"context"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/std/errors"
)

var _ database.Database[struct{}] = (*clickHouse[struct{}])(nil)

func New[T any](conn clickhouse.Conn) database.Database[T] {
	return &clickHouse[T]{
		conn: conn,
	}
}

type clickHouse[T any] struct {
	conn clickhouse.Conn
}

// BeginTx implements database.Database.
func (c *clickHouse[T]) BeginTx(
	ctx context.Context) (database.Transaction[T], error) {
	return nil, errors.F("[sentinez] clickhouse transaction not supported")
}

// Collect implements database.Database.
func (c *clickHouse[T]) CollectRows(ctx context.Context,
	builder database.SQLBuilder,
	scan func(database.Rows) ([]T, error)) ([]T, error) {

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scan(rows)
}

func (c *clickHouse[T]) Collect(ctx context.Context,
	builder database.SQLBuilder,
	scan func(database.Row) (*T, error)) (*T, error) {

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scan(rows)
}

// Exec implements database.Database.
func (c *clickHouse[T]) Exec(ctx context.Context,
	builder database.SQLBuilder) (database.ExecResult, error) {

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	if err := c.conn.Exec(ctx, query, args...); err != nil {
		return nil, err
	}

	return nil, nil
}

// WithTx implements database.Database.
func (c *clickHouse[T]) WithTx(ctx context.Context,
	fn func(database.Transaction[T]) error) error {

	return errors.F("[sentinez] clickhouse transaction not supported")
}
