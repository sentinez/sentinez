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
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/pkg/auto/queries/users"
	"github.com/sentinez/sentinez/pkg/common/uuid"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/infra/database/postgresz"
	"github.com/sentinez/sentinez/pkg/infra/database/query"
	"github.com/sentinez/sentinez/pkg/std/table"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	_ database.Repository[*iam.Users, string] = (*Users)(nil)
	_ IUser                                   = (*Users)(nil)
)

type IUser interface {
	Create(ctx context.Context, user *iam.Users) (*iam.Users, error)
	Update(ctx context.Context, user *iam.Users) (*iam.Users, error)
	Get(ctx context.Context, id string) (*iam.Users, error)
	Delete(ctx context.Context, id string) error

	// extra methods

	GetByUsernameOrEmail(ctx context.Context,
		input string) (*iam.Users, error)

	List(ctx context.Context,
		req *iam.ListUsersRequest) (*iam.ListUsersResponse, error)

	Total(ctx context.Context, req *iam.ListUsersRequest) (int64, error)
}

func New(pool *pgxpool.Pool) (IUser, error) {
	tableName := table.Table(table.Users)

	storage, err := postgresz.New[*iam.Users](pool, tableName,
		postgresz.WithIndex("email", "phone_number", "username"))
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

// GetByUsernameOrEmail implements IUser.
func (u *Users) GetByUsernameOrEmail(ctx context.Context,
	input string) (*iam.Users, error) {

	user, err := u.query.GetByUsernameOrEmail(ctx, []byte(input))
	if err != nil {
		return nil, err
	}

	var result iam.Users
	if err := protojson.Unmarshal(user.Data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// nolint:funlen
func buildListQuery(builder sq.SelectBuilder,
	req *iam.ListUsersRequest) sq.SelectBuilder {
	for _, id := range req.GetIds() {
		builder = builder.Where(sq.Eq{
			postgresz.Primary(iam.UsersFieldId): id,
		})
	}

	for _, email := range req.GetEmails() {
		builder = builder.Where(sq.Eq{
			postgresz.Field(iam.UsersFieldEmail): email,
		})
	}

	for _, phone := range req.GetPhoneNumbers() {
		builder = builder.Where(sq.Eq{
			postgresz.Field(iam.UsersFieldPhoneNumber): phone,
		})
	}

	return builder
}

// List implements IUser.
// nolint:funlen
func (u *Users) List(ctx context.Context,
	req *iam.ListUsersRequest) (*iam.ListUsersResponse, error) {

	builder := sq.Select(database.Data).From(u.tableName)
	builder = query.Paging(builder, req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	zlog.Debug("[users] query: ", query, " args: ", args)

	users, err := u.storage.CollectRows(
		ctx, builder, postgresz.Scans[*iam.Users])
	if err != nil {
		return nil, err
	}

	if req.GetPage().GetTotal() {
		total, err = u.Total(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	return &iam.ListUsersResponse{Users: users, Total: total}, nil
}

func (u *Users) Total(ctx context.Context,
	req *iam.ListUsersRequest) (int64, error) {

	builder := sq.Select("COUNT(*) AS count").From(u.tableName)
	builder = buildListQuery(builder, req)

	var count int64
	err := u.storage.Query(ctx, builder, &count)

	return count, err
}

// Create implements IUser.
func (u *Users) Create(ctx context.Context,
	user *iam.Users) (*iam.Users, error) {

	user.Id = uuid.Generate(u.tableName)

	data, err := protojson.Marshal(user)
	if err != nil {
		return nil, err
	}

	_, err = u.query.Insert(ctx, users.InsertParams{
		ID:      user.GetId(),
		Column2: data,
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Delete implements IUser.
func (u *Users) Delete(ctx context.Context, id string) error {
	return u.query.Delete(ctx, id)
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
