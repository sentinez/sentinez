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

	"github.com/jackc/pgx/v5"
	iammodel "github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/models/v1"
	"github.com/sentinez/sentinez/mods/queries/gen/users"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/infra/database/sql"
	"github.com/sentinez/sentinez/pkg/std/table"
)

var (
	_ database.Repository[*iammodel.Users, string] = (*Users)(nil)
	_ IUser                                        = (*Users)(nil)
)

type IUser interface {
	Create(ctx context.Context, user *iammodel.Users) (*iammodel.Users, error)
	Update(ctx context.Context,
		id string, user *iammodel.Users) (*iammodel.Users, error)
	Get(ctx context.Context, id string) (*iammodel.Users, error)
	GetAll(ctx context.Context) ([]*iammodel.Users, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
	Count(ctx context.Context) (int64, error)
}

func New(pgCon *pgx.Conn) IUser {
	storage := sql.New[*iammodel.Users, string](pgCon, table.Users)

	return &Users{
		SQL:   storage,
		query: users.New(pgCon),
	}
}

type Users struct {
	*sql.SQL[*iammodel.Users, string]
	query *users.Queries
}
