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

package usersrepo

import (
	"context"

	sql "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/pkg/auto/queries/gen/users"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/infra/database/postgresdb"
	pgopt "github.com/sentinez/sentinez/pkg/infra/options/postgres"
	"github.com/sentinez/sentinez/pkg/std/table"
)

var (
	_ database.Repository[*iam.Users, string] = (*Users)(nil)
	_ IUser                                   = (*Users)(nil)
)

type IUser interface {
	Create(ctx context.Context, user *iam.Users) (*iam.Users, error)
	Update(ctx context.Context,
		id string, user *iam.Users) (*iam.Users, error)
	Get(ctx context.Context, id string) (*iam.Users, error)
	GetAll(ctx context.Context) ([]*iam.Users, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
	Count(ctx context.Context) (int64, error)
}

func New(pool *pgxpool.Pool) (IUser, error) {
	tableName := table.Table(table.Users)

	storage, err := postgresdb.New[*iam.Users](pool, tableName, pgopt.TableKV)
	if err != nil {
		return nil, err
	}

	return &Users{
		query:   users.New(pool),
		storage: storage,
	}, nil
}

type Users struct {
	query   *users.Queries
	storage database.Database[*iam.Users]
}

// Count implements IUser.
func (u *Users) Count(ctx context.Context) (int64, error) {
	_ = ctx
	panic("unimplemented")
}

// Create implements IUser.
func (u *Users) Create(ctx context.Context,
	user *iam.Users) (*iam.Users, error) {

	_, _ = ctx, user

	panic("unimplemented")
}

// Delete implements IUser.
func (u *Users) Delete(ctx context.Context, id string) error {
	_, _ = ctx, id

	panic("unimplemented")
}

// Exists implements IUser.
func (u *Users) Exists(ctx context.Context, id string) (bool, error) {
	_, _ = ctx, id

	panic("unimplemented")
}

// Get implements IUser.
func (u *Users) Get(ctx context.Context, id string) (*iam.Users, error) {
	_, _ = ctx, id

	panic("unimplemented")
}

// GetAll implements IUser.
func (u *Users) GetAll(ctx context.Context) ([]*iam.Users, error) {
	query := sql.Select("*").From("users")

	return u.storage.CollectRows(ctx, query, scan)
}

// Update implements IUser.
func (u *Users) Update(ctx context.Context,
	id string, user *iam.Users) (*iam.Users, error) {

	_, _, _ = ctx, id, user
	panic("unimplemented")
}

func scan(rows database.Rows) ([]*iam.Users, error) {
	var users []*iam.Users

	for rows.Next() {
		var user iam.Users
		if err := rows.Scan(); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
