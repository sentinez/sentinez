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
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func protoKindToSQL(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.StringKind:
		return "TEXT"
	case protoreflect.Int32Kind,
		protoreflect.Sint32Kind,
		protoreflect.Uint32Kind:
		return "INTEGER"
	case protoreflect.Int64Kind,
		protoreflect.Sint64Kind,
		protoreflect.Uint64Kind:
		return "BIGINT"
	case protoreflect.BoolKind:
		return "BOOLEAN"
	case protoreflect.DoubleKind,
		protoreflect.FloatKind:
		return "DOUBLE PRECISION"
	case protoreflect.BytesKind:
		return "BYTEA"
	default:
		return "JSONB"
	}
}

func columnExists(ctx context.Context,
	pool *pgxpool.Pool, table, column string) (bool, error) {

	const query = `
	SELECT 1 FROM information_schema.columns
	WHERE table_name = $1 AND column_name = $2 LIMIT 1;
	`
	row := pool.QueryRow(ctx, query, table, column)
	var dummy int
	err := row.Scan(&dummy)
	if err != nil {
		return false, nil // not exists
	}

	return true, nil
}

func tableExists(ctx context.Context,
	pool *pgxpool.Pool, table string) (bool, error) {

	const query = `
	SELECT 1 FROM information_schema.tables
	WHERE table_name = $1 LIMIT 1;
	`
	row := pool.QueryRow(ctx, query, table)
	var dummy int
	err := row.Scan(&dummy)
	if err != nil {
		return false, nil
	}

	return true, nil
}

//nolint:funlen
func syncProtoToPostgres(ctx context.Context,
	pool *pgxpool.Pool, table string, msg proto.Message) error {

	desc := msg.ProtoReflect().Descriptor()
	fields := desc.Fields()

	exists, err := tableExists(ctx, pool, table)
	if err != nil {
		return fmt.Errorf("check table exists: %w", err)
	}

	if !exists {

		var cols []string
		for i := 0; i < fields.Len(); i++ {
			f := fields.Get(i)
			colName := string(f.Name())
			sqlType := protoKindToSQL(f.Kind())
			cols = append(cols, fmt.Sprintf("%s %s", colName, sqlType))
		}

		ddl := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s);`,
			table, strings.Join(cols, ", "))
		_, err := pool.Exec(ctx, ddl)
		if err != nil {
			return fmt.Errorf("create table failed: %w", err)
		}

	} else {

		for i := 0; i < fields.Len(); i++ {
			f := fields.Get(i)
			colName := string(f.Name())
			sqlType := protoKindToSQL(f.Kind())

			ok, err := columnExists(ctx, pool, table, colName)
			if err != nil {
				return fmt.Errorf("check column exists: %w", err)
			}
			if !ok {
				_, err := pool.Exec(ctx,
					fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s;`,
						table, colName, sqlType))
				if err != nil {
					return fmt.Errorf("alter table failed: %w", err)
				}
			}
		}
	}

	return nil
}
