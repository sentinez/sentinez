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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	modelpb "github.com/sentinez/sentinez/api/gen/go/sentinez/std/model/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/pkg/common/uuid"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/infra/database/postgres"
	"github.com/sentinez/sentinez/pkg/infra/table"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	_ IUser = (*Users)(nil)
)

type IUser interface {
	Create(ctx context.Context, user *iam.Users) (*iam.Users, error)
	Update(ctx context.Context, user *iam.Users) (*iam.Users, error)
	Get(ctx context.Context, id string) (*iam.Users, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IUser

	// extra methods

	GetByUsernameOrEmail(ctx context.Context,
		input string) (*iam.Users, error)

	List(ctx context.Context,
		req *iam.ListUsersRequest) (*iam.ListUsersResponse, error)

	Total(ctx context.Context, req *iam.ListUsersRequest) (int64, error)
}

func New(conf *common.Config) (IUser, error) {
	tableName := table.NewTable(table.Users)

	storage, err := postgres.New[*iam.Users](conf, tableName,
		postgres.WithIndex("email", "phone_number", "username"))
	if err != nil {
		return nil, err
	}

	return &Users{
		storage:   storage,
		tableName: tableName,
	}, nil
}

type Users struct {
	tableName string
	storage   database.Database[*iam.Users]
}

func (u *Users) WithTX(tx *postgres.TxSession) IUser {
	return &Users{
		tableName: u.tableName,
		storage:   postgres.WithTx[*iam.Users](tx, u.tableName),
	}
}

// GetByUsernameOrEmail implements IUser.
func (u *Users) GetByUsernameOrEmail(ctx context.Context,
	input string) (*iam.Users, error) {

	builder := sq.Select(database.SchemalessFieldData).From(u.tableName).Where(
		sq.Or{
			sq.Eq{postgres.Field(iam.UsersFieldFullName): input},
			sq.Eq{postgres.Field(iam.UsersFieldEmail): input},
		},
	)

	return u.storage.CollectOneRow(ctx, builder, postgres.Scan[*iam.Users])
}

// nolint:funlen
func buildListQuery(builder sq.SelectBuilder,
	req *iam.ListUsersRequest) sq.SelectBuilder {
	for _, id := range req.GetIds() {
		builder = builder.Where(sq.Eq{
			postgres.Primary(iam.UsersFieldId): id,
		})
	}

	for _, email := range req.GetEmails() {
		builder = builder.Where(sq.Eq{
			postgres.Field(iam.UsersFieldEmail): email,
		})
	}

	for _, phone := range req.GetPhoneNumbers() {
		builder = builder.Where(sq.Eq{
			postgres.Field(iam.UsersFieldPhoneNumber): phone,
		})
	}

	return builder
}

// List implements IUser.
// nolint:funlen
func (u *Users) List(ctx context.Context,
	req *iam.ListUsersRequest) (*iam.ListUsersResponse, error) {

	builder := sq.Select(database.SchemalessFieldData).From(u.tableName)
	builder = postgres.Paging(builder, req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	zlog.Debug("[users] query: ", query, " args: ", args)

	users, err := u.storage.CollectRows(
		ctx, builder, postgres.Scans[*iam.Users])
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

	now := timestamppb.Now()
	user.Id = uuid.NewID(table.NewPrimaryKey(table.Users))
	user.Metadata = &modelpb.Metadata{
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       user.GetMetadata().GetCreatedBy(),
		UpdatedBy:       user.GetMetadata().GetUpdatedBy(),
		ResourceOwnerId: user.GetId(),
	}

	if err := u.storage.Set(ctx, user.GetId(), user); err != nil {
		return nil, err
	}

	return user, nil
}

// Delete implements IUser.
func (u *Users) Delete(ctx context.Context, id string) error {
	return u.storage.Delete(ctx, id)
}

// Get implements IUser.
func (u *Users) Get(ctx context.Context, id string) (*iam.Users, error) {
	return u.storage.Get(ctx, id)
}

// Update implements IUser.
func (u *Users) Update(
	ctx context.Context, user *iam.Users) (*iam.Users, error) {

	user.Metadata.UpdatedAt = timestamppb.Now()
	if err := u.storage.Set(ctx, user.GetId(), user); err != nil {
		return nil, err
	}

	return user, nil
}
