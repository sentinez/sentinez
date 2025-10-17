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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	commonpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	modelpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/model/v1"
	"github.com/sentinez/sentinez/internal/shared/tables"
	"github.com/sentinez/sentinez/pkg/storage/database"
	"github.com/sentinez/sentinez/pkg/storage/database/postgres"
	"github.com/sentinez/sentinez/pkg/storage/utils/table"
	"github.com/sentinez/sentinez/pkg/x/uuidx"
	"github.com/sentinez/sentinez/pkg/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ IUser = (*Users)(nil)
)

type IUser interface {
	Create(ctx context.Context, user *iam.User) (*iam.User, error)
	Update(ctx context.Context, user *iam.User) (*iam.User, error)
	Get(ctx context.Context, id string) (*iam.User, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IUser

	// extra methods

	GetByFullnameOrEmail(ctx context.Context, input string) (*iam.User, error)

	List(ctx context.Context,
		req *iam.ListUsersRequest) (*iam.ListUsersResponse, error)

	Total(ctx context.Context, req *iam.ListUsersRequest) (int64, error)
}

func New(appConf *commonpb.AppConfig) (IUser, error) {
	storage, err := postgres.New[*iam.User](appConf, tables.Users)
	if err != nil {
		return nil, err
	}

	return &Users{
		storage: storage,
	}, nil
}

type Users struct {
	storage database.Database[*iam.User]
}

func (u *Users) WithTX(tx *postgres.TxSession) IUser {
	return &Users{
		storage: postgres.WithTx(tx, u.storage),
	}
}

// GetByFullnameOrEmail implements IUser.
func (u *Users) GetByFullnameOrEmail(ctx context.Context,
	input string) (*iam.User, error) {

	builder := postgres.SelectBuilder(u.storage, nil)
	builder = builder.From(u.storage.Table()).
		Where(sq.Or{
			sq.Eq{postgres.Field(iam.UserFieldFullName): input},
			sq.Eq{postgres.Field(iam.UserFieldEmail): input},
		})

	return u.storage.CollectOneRow(ctx, builder, postgres.Scan[*iam.User])
}

// nolint:funlen
func buildListQuery(builder sq.SelectBuilder,
	req *iam.ListUsersRequest) sq.SelectBuilder {
	for _, id := range req.GetIds() {
		builder = builder.Where(sq.Eq{
			postgres.Primary(iam.UserFieldId): id,
		})
	}

	for _, email := range req.GetEmails() {
		builder = builder.Where(sq.Eq{
			postgres.Field(iam.UserFieldEmail): email,
		})
	}

	for _, phone := range req.GetPhoneNumbers() {
		builder = builder.Where(sq.Eq{
			postgres.Field(iam.UserFieldPhoneNumber): phone,
		})
	}

	return builder
}

// List implements IUser.
// nolint:funlen
func (u *Users) List(ctx context.Context,
	req *iam.ListUsersRequest) (*iam.ListUsersResponse, error) {

	builder := postgres.SelectBuilder(u.storage, req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	zlog.Debug("[users] query: ", query, " args: ", args)

	users, err := u.storage.CollectRows(
		ctx, builder, postgres.Scans[*iam.User])
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

	builder := sq.
		Select("COUNT(*) AS count").
		From(u.storage.Table())

	builder = buildListQuery(builder, req)

	var count int64
	err := u.storage.Query(ctx, builder, &count)

	return count, err
}

// Create implements IUser.
func (u *Users) Create(ctx context.Context,
	user *iam.User) (*iam.User, error) {

	now := timestamppb.Now()
	user.Id = uuidx.NewID(table.NewPrimaryKey(tables.Users))
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
func (u *Users) Get(ctx context.Context, id string) (*iam.User, error) {
	return u.storage.Get(ctx, id)
}

// Update implements IUser.
func (u *Users) Update(
	ctx context.Context, user *iam.User) (*iam.User, error) {

	user.Metadata.UpdatedAt = timestamppb.Now()
	if err := u.storage.Set(ctx, user.GetId(), user); err != nil {
		return nil, err
	}

	return user, nil
}
