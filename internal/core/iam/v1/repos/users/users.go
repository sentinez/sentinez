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
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	modelpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/model/v1"
	"github.com/sentinez/sentinez/internal/shared/tables"
	"github.com/sentinez/sentinez/pkg/storage/dbx"
	"github.com/sentinez/sentinez/pkg/storage/dbx/postgres"
	"github.com/sentinez/sentinez/pkg/storage/utils/table"
	"github.com/sentinez/shared/ids"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ IUser = (*Users)(nil)
)

// nolint
type IUser interface {
	Create(ctx context.Context, user *iam.User) (*iam.User, error)
	Update(ctx context.Context, user *iam.User) error
	Get(ctx context.Context, id string) (*iam.User, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IUser

	// extra methods

	GetByFullnameOrEmail(ctx context.Context, input string) (*iam.User, error)
	List(ctx context.Context, req *iam.ListUsersRequest) (*iam.ListUsersResponse, error)
	Total(ctx context.Context, req *iam.ListUsersRequest) (int64, error)
}

func New(ctx context.Context, appConf *confpb.Config) (IUser, error) {
	storage, err := postgres.New[iam.User](ctx, appConf,
		dbx.WithTable(tables.Users),
		dbx.WithColumn(iam.User_Id, postgres.String),
		dbx.WithColumn(iam.User_Email, postgres.String),
		dbx.WithColumn(iam.User_PhoneNumber, postgres.String),
		dbx.WithColumn(iam.User_FullName, postgres.String),
	)
	if err != nil {
		return nil, err
	}

	return &Users{
		storage: storage,
	}, nil
}

type Users struct {
	storage dbx.Database[iam.User]
}

func (u *Users) WithTX(tx *postgres.TxSession) IUser {
	return &Users{
		storage: postgres.WithTx(tx, u.storage),
	}
}

// GetByFullnameOrEmail implements IUser.
func (u *Users) GetByFullnameOrEmail(ctx context.Context,
	input string) (*iam.User, error) {

	builder := u.selectQuery(nil)
	builder = builder.From(u.storage.Table()).
		Where(sq.Or{
			sq.Eq{iam.User_FullName: input},
			sq.Eq{iam.User_Email: input},
		})

	return u.storage.CollectOneRow(ctx, builder, scanOne)
}

// nolint:funlen
func buildListQuery(builder sq.SelectBuilder,
	req *iam.ListUsersRequest) sq.SelectBuilder {
	for _, id := range req.GetIds() {
		builder = builder.Where(sq.Eq{iam.User_Id: id})
	}

	for _, email := range req.GetEmails() {
		builder = builder.Where(sq.Eq{iam.User_Email: email})
	}

	for _, phone := range req.GetPhoneNumbers() {
		builder = builder.Where(sq.Eq{iam.User_PhoneNumber: phone})
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

	users, err := u.storage.CollectRows(ctx, builder, scan)
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
func (u *Users) Create(ctx context.Context, user *iam.User) (*iam.User, error) {
	user.Id = ids.NewID(table.NewPrimaryKey(tables.Users))
	query := postgres.InsertBuilder(u.storage,
		[]string{
			iam.User_Id,
			iam.User_Email,
			iam.User_FullName,
			iam.User_PhoneNumber},
		[]any{
			user.GetId(),
			user.GetEmail(),
			user.GetFullName(),
			user.GetPhoneNumber(),
		},
	)

	if _, err := u.storage.Insert(ctx, query); err != nil {
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
	builder := u.selectQuery(nil).Where(sq.Eq{iam.User_Id: id})
	return u.storage.Select(ctx, builder, scanOne)
}

// Update implements IUser.
func (u *Users) Update(ctx context.Context, user *iam.User) error {

	query := postgres.UpdateBuilder(u.storage, user.GetId())

	if user.GetEmail() != "" {
		query = query.Set(iam.User_Email, user.GetEmail())
	}

	if user.GetFullName() != "" {
		query = query.Set(iam.User_FullName, user.GetFullName())
	}

	if user.GetPhoneNumber() != "" {
		query = query.Set(iam.Account_Password, user.GetPhoneNumber())
	}

	_, err := u.storage.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (u *Users) selectQuery(page *common.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(u.storage, page,
		iam.User_Id,
		iam.User_Email,
		iam.User_FullName,
		iam.User_PhoneNumber,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*iam.User, error) {
	var results []*iam.User

	for rows.Next() {
		var (
			createdAt, updatedAt time.Time
			user                 iam.User
		)
		err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.FullName,
			&user.PhoneNumber,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		user.Metadata = &modelpb.Metadata{
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: timestamppb.New(updatedAt),
		}

		results = append(results, &user)
	}

	return results, rows.Err()
}

func scanOne(row dbx.Row) (*iam.User, error) {
	var (
		createdAt, updatedAt time.Time
		user                 iam.User
	)
	err := row.Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
		&user.PhoneNumber,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.Metadata = &modelpb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &user, nil
}
