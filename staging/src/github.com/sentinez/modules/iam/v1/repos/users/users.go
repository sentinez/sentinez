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
	"github.com/sentinez/core/common/tables"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/postgres"
	"github.com/sentinez/core/storage/utils/table"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/modules/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ IUser = (*Users)(nil)
)

// nolint
type IUser interface {
	Create(ctx context.Context, user *iampb.User) (*iampb.User, error)
	Update(ctx context.Context, user *iampb.User) error
	Get(ctx context.Context, id string) (*iampb.User, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IUser

	// extra methods

	List(ctx context.Context, req *iampb.ListUsersRequest) (*iampb.ListUsersResponse, error)
}

func New(ctx context.Context, appConf *settingpb.Config) (IUser, error) {
	storage, err := postgres.New[iampb.User](ctx, appConf,
		dbx.WithTable(tables.IAMUsers),
		dbx.WithColumns(dbx.ColumnM{
			iampb.User_Id:       postgres.String,
			iampb.User_Email:    postgres.String,
			iampb.User_FullName: postgres.String,
			iampb.User_Console:  postgres.Int4,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Users{
		storage: storage,
	}, nil
}

type Users struct {
	storage dbx.Database[iampb.User]
}

func (u *Users) WithTX(tx *postgres.TxSession) IUser {
	return &Users{
		storage: postgres.WithTx(tx, u.storage),
	}
}

// nolint:funlen
func buildListQuery(builder sq.SelectBuilder,
	req *iampb.ListUsersRequest) sq.SelectBuilder {
	for _, id := range req.GetIds() {
		builder = builder.Where(sq.Eq{iampb.User_Id: id})
	}

	for _, email := range req.GetEmails() {
		builder = builder.Where(sq.Eq{iampb.User_Email: email})
	}

	return builder
}

// List implements IUser.
// nolint:funlen
func (u *Users) List(ctx context.Context,
	req *iampb.ListUsersRequest) (*iampb.ListUsersResponse, error) {

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
		total, err = u.storage.Total(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &iampb.ListUsersResponse{Users: users, Total: total}, nil
}

// Create implements IUser.
func (u *Users) Create(ctx context.Context,
	user *iampb.User) (*iampb.User, error) {

	user.Id = rand.NewID(table.NewPrimaryKey(tables.IAMUsers))
	query := postgres.InsertBuilder(u.storage, postgres.M{
		iampb.User_Id:       user.GetId(),
		iampb.User_Email:    user.GetEmail(),
		iampb.User_FullName: user.GetFullName(),
		iampb.User_Console:  user.GetConsole(),
	})

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
func (u *Users) Get(ctx context.Context, id string) (*iampb.User, error) {
	builder := u.selectQuery(nil).Where(sq.Eq{iampb.User_Id: id})
	return u.storage.Select(ctx, builder, scanOne)
}

// Update implements IUser.
func (u *Users) Update(ctx context.Context, user *iampb.User) error {

	query := postgres.UpdateBuilder(u.storage, user.GetId())

	if user.GetEmail() != "" {
		query = query.Set(iampb.User_Email, user.GetEmail())
	}

	if user.GetFullName() != "" {
		query = query.Set(iampb.User_FullName, user.GetFullName())
	}

	_, err := u.storage.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (u *Users) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(u.storage, page,
		iampb.User_Id,
		iampb.User_Email,
		iampb.User_FullName,
		iampb.User_Console,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*iampb.User, error) {
	var users []*iampb.User

	for rows.Next() {
		user, err := scanOne(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, rows.Err()
}

func scanOne(row dbx.Row) (*iampb.User, error) {
	var (
		createdAt, updatedAt time.Time
		user                 iampb.User
	)

	err := row.Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
		&user.Console,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.Metadata = &typepb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &user, nil
}
