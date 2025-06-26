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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sentinez/sentinez/pkg/infra/database"
)

var _ database.Transaction[struct{}] = (*PostgresTx[struct{}])(nil)

type PostgresTx[T any] struct {
	tx pgx.Tx
}

// Collect implements database.Transaction.
func (p *PostgresTx[T]) CollectRows(ctx context.Context,
	builder database.SQLBuilder,
	_ func(database.Rows) ([]T, error)) ([]T, error) {

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
func (p *PostgresTx[T]) Commit(ctx context.Context) error {
	if err := p.tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// Exec implements database.Transaction.
func (p *PostgresTx[T]) Exec(ctx context.Context,
	builder database.SQLBuilder) (database.ExecResult, error) {

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

// Rollback implements database.Transaction.
func (p *PostgresTx[T]) Rollback(ctx context.Context) error {
	if err := p.tx.Rollback(ctx); err != nil {
		return err
	}

	return nil
}
