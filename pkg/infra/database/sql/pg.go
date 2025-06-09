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

// Package sql provides an implementation of the database using PostgreSQL.
package sql

import (
	"context"
	"errors"
	"fmt"

	celserr "github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/table"

	"github.com/jackc/pgx/v5"
)

// SQL is a generic repository for PostgresSQL using pgx.
type SQL[T any, ID comparable] struct {
	db        *pgx.Conn
	tableName string
}

// New initializes a new SQLStorage with PostgresSQL.
// SQLStorage is a generic repository for PostgresSQL using pgx.
// Must be implemented Create method by the user.
func New[T any, ID comparable](
	db *pgx.Conn, tableName string) *SQL[T, ID] {

	// wrap the table name with the current mode
	// e.g. "dev.users" or "prod.users"
	tableName = table.Wrapper(tableName)

	return &SQL[T, ID]{db: db, tableName: tableName}
}

// Create inserts a new record into the database.
func (s *SQL[T, ID]) Create(ctx context.Context, entity T) (T, error) {

	_ = ctx
	_ = entity

	return entity, celserr.ErrUnimplemented
}

// Get retrieves a record by ID.
func (s *SQL[T, ID]) Get(ctx context.Context, id ID) (T, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", s.tableName)
	row := s.db.QueryRow(ctx, query, id)

	var entity T
	if err := row.Scan(&entity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity, fmt.Errorf("record not found")
		}

		return entity, err
	}

	return entity, nil
}

// GetAll retrieves all records from the table.
func (s *SQL[T, ID]) GetAll(ctx context.Context) ([]T, error) {
	query := fmt.Sprintf("SELECT * FROM %s", s.tableName)
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []T
	for rows.Next() {
		var entity T
		if err := rows.Scan(&entity); err != nil {
			return nil, err
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

// Update modifies a record by ID.
func (s *SQL[T, ID]) Update(ctx context.Context, id ID, entity T) (T, error) {
	_ = ctx
	_ = id
	_ = entity

	return entity, celserr.ErrUnimplemented
}

// Delete removes a record by ID.
func (s *SQL[T, ID]) Delete(ctx context.Context, id ID) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", s.tableName)
	_, err := s.db.Exec(ctx, query, id)

	return err
}

// Exists checks if a record exists by ID.
func (s *SQL[T, ID]) Exists(ctx context.Context, id ID) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1", s.tableName)
	row := s.db.QueryRow(ctx, query, id)

	var exists int
	err := row.Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	return err == nil, err
}

// Count returns the total number of records in the table.
func (s *SQL[T, ID]) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", s.tableName)
	row := s.db.QueryRow(ctx, query)

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
