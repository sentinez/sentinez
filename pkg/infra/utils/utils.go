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

// Package utils provides utility functions for the service.
package utils

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
)

// NewPgxPool create new pool connection for multiple query
func NewPgxPool(conf *common.Config) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), conf.GetPostgresUri())
}

// NewPgxConn create new connection for single query
func NewPgxConn(
	ctx context.Context, conf *common.Config) (*pgx.Conn, error) {

	_ = conf
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		"conf.DbUser",
		"conf.DbPassword",
		"conf.DbHost",
		"conf.DbPort",
		"conf.DbName",
	)

	return pgx.Connect(ctx, dsn)
}

// Filter is a function that filters a slice of object generically
func Filter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}
