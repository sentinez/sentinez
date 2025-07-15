// Copyright 2025 Sentinez Labs.
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

	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Reference(pool *pgxpool.Pool,
	fromTable, fromField, toTable, toField string) error {
	ctx := context.Background()
	constraintName := fmt.Sprintf("fk_%s_%s", fromTable, fromField)

	var exists int
	err := pool.QueryRow(ctx, checkConstraint, constraintName).Scan(&exists)
	if err == nil {

		zlog.Debugf("[pgopt] foreign key already exists: %s (%s.%s -> %s.%s)",
			constraintName, fromTable, fromField, toTable, toField)
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {

		zlog.Errorf("[pgopt] check constraint error: %v", err)
		return errors.F("[pgopt] failed to check constraint existence: %w", err)
	}

	sql := fmt.Sprintf(alterQuery,
		fromTable, constraintName, fromField, toTable, toField)

	_, err = pool.Exec(ctx, sql)
	if err != nil {
		zlog.Errorf("[pgopt] executing ALTER TABLE error: %v", err)
		return errors.F("[pgopt] failed to execute ALTER TABLE: %w", err)
	}

	zlog.Debugf("[pgopt] reference created: %s (%s.%s -> %s.%s)",
		constraintName, fromTable, fromField, toTable, toField)

	return nil
}

func WithReference(fromField, toTable, toField string) database.Option {
	return func(ref *database.Table) {
		if ref.References == nil {
			ref.References = make(map[string]database.Reference)
		}

		ref.References[fromField] = database.Reference{
			FromField: fromField,
			ToTable:   toTable,
			ToField:   toField,
		}
	}
}
