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

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/pkg/auto/queries/users"
	"github.com/sentinez/sentinez/pkg/common/uuid"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/infra/database/postgresdb"
	pgopt "github.com/sentinez/sentinez/pkg/infra/options/postgres"
	"github.com/sentinez/sentinez/pkg/std/table"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	_ database.Repository[*iam.Users, string] = (*Users)(nil)
	_ IUser                                   = (*Users)(nil)
)

type IUser interface {
	Create(ctx context.Context, user *iam.Users) (*iam.Users, error)
	Update(ctx context.Context, user *iam.Users) (*iam.Users, error)
	Get(ctx context.Context, id string) (*iam.Users, error)
	GetMany(ctx context.Context, page *common.Pages) ([]*iam.Users, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
	Count(ctx context.Context) (int64, error)

	// extra methods
	List(ctx context.Context, req *iam.ListUsersRequest) ([]*iam.Users, error)
}

func New(pool *pgxpool.Pool) (IUser, error) {
	tableName := table.Table(table.Users)

	storage, err := postgresdb.New[*iam.Users](pool, tableName,
		pgopt.WithStorageOption(database.StorageKV),
	)
	if err != nil {
		return nil, err
	}

	return &Users{
		query:     users.New(pool),
		storage:   storage,
		tableName: tableName,
	}, nil
}

type Users struct {
	tableName string
	query     *users.Queries
	storage   database.Database[*iam.Users]
}

// List implements IUser.
func (u *Users) List(ctx context.Context,
	req *iam.ListUsersRequest) ([]*iam.Users, error) {

	builder := sq.Select(pgopt.ColumnData).From(u.tableName)

	for _, id := range req.GetIds() {
		builder = builder.Where(sq.Eq{"id": id})
	}

	for _, email := range req.GetEmails() {
		builder = builder.Where(sq.Eq{"data->>'email'": email})
	}

	for _, phone := range req.GetPhoneNumbers() {
		builder = builder.Where(sq.Eq{"data->>'phone_number'": phone})
	}

	return u.storage.CollectRows(ctx, builder, scan)
}

func scan(r database.Rows) ([]*iam.Users, error) {
	var users []*iam.Users

	for r.Next() {
		var (
			data []byte
		)
		if err := r.Scan(&data); err != nil {
			return nil, err
		}

		var user iam.Users
		if err := protojson.Unmarshal(data, &user); err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err := r.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Count implements IUser.
func (u *Users) Count(ctx context.Context) (int64, error) {
	return u.query.Count(ctx)
}

// Create implements IUser.
func (u *Users) Create(ctx context.Context,
	user *iam.Users) (*iam.Users, error) {

	user.Id = uuid.Generate(table.Users)

	data, err := protojson.Marshal(user)
	if err != nil {
		return nil, err
	}

	_, err = u.query.Insert(ctx, data)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Delete implements IUser.
func (u *Users) Delete(ctx context.Context, id string) error {
	return u.query.Delete(ctx, id)
}

// Exists implements IUser.
func (u *Users) Exists(ctx context.Context, id string) (bool, error) {
	user, err := u.query.GetByID(ctx, id)
	if err != nil || user.ID == "" {
		return false, err
	}

	return true, nil
}

// Get implements IUser.
func (u *Users) Get(ctx context.Context, id string) (*iam.Users, error) {
	user, err := u.query.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var result iam.Users
	if err := protojson.Unmarshal(user.Data, &result); err != nil {
		return nil, err
	}

	return &result, err
}

// GetMany implements IUser.
func (u *Users) GetMany(ctx context.Context,
	page *common.Pages) ([]*iam.Users, error) {

	offset := database.GetOffset(int(page.GetIndex()), int(page.GetSize()))
	result, err := u.query.GetPage(ctx, users.GetPageParams{
		Offset: int32(offset),
		Limit:  int32(page.GetSize()),
	})
	if err != nil {
		return nil, err
	}

	var resp []*iam.Users
	for _, user := range result {
		var u iam.Users
		if err := protojson.Unmarshal(user.Data, &u); err != nil {
			return nil, err
		}

		u.Id = user.ID
		resp = append(resp, &u)
	}

	return resp, nil
}

// Update implements IUser.
func (u *Users) Update(
	ctx context.Context, user *iam.Users) (*iam.Users, error) {

	data, err := protojson.Marshal(user)
	if err != nil {
		return nil, err
	}

	err = u.query.Update(ctx,
		users.UpdateParams{ID: user.GetId(), Column1: data})
	if err != nil {
		return nil, err
	}

	return user, nil
}
