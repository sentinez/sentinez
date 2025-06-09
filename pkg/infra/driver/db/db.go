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

// Package db implement database driver
package db

import (
	"context"
)

// Driver is an interface for database driver.
type Driver[T any] interface {
	Query(ctx context.Context, sql string, args ...any) (Rows[T], error)
}

// Rows is an interface for database rows.
type Rows[T any] interface {
	CollectOne() (T, error)
	CollectAll() ([]T, error)
}
