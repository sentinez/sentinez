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
	"github.com/sentinez/core/szcmn/sztables"
	"github.com/sentinez/core/szstorage/szdbx"
	"github.com/sentinez/core/szstorage/szdbx/szpostgres"
	"github.com/sentinez/core/szstorage/szstorageutils/sztable"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/errorx"
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

	WithTX(tx *szpostgres.TxSession) IUser

	// extra methods

	GetByEmail(ctx context.Context, email string) (*iampb.User, error)
	List(ctx context.Context, req *iampb.ListUsersRequest) (*iampb.ListUsersResponse, error)
}

func New(ctx context.Context, appConf *settingpb.Config) (IUser, error) {
	storage, err := szpostgres.New[iampb.User](ctx, appConf,
		szdbx.WithTable(sztables.IAMUsers),
		szdbx.WithColumns(szdbx.ColumnM{
			iampb.User_Id:           szpostgres.String,
			iampb.User_Email:        szpostgres.String,
			iampb.User_FullName:     szpostgres.String,
			iampb.User_ControlPlane: szpostgres.Int4,
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
	storage szdbx.Database[iampb.User]
}

func (u *Users) WithTX(tx *szpostgres.TxSession) IUser {
	return &Users{
		storage: szpostgres.WithTx(tx, u.storage),
	}
}

// listConditions maps the filters of the request to WHERE conditions. A
// user matches a filter when it has any of the listed values (IN), and
// it must match every filter that is set.
func listConditions(req *iampb.ListUsersRequest) []sq.Sqlizer {
	var conds []sq.Sqlizer

	if ids := req.GetIds(); len(ids) != 0 {
		conds = append(conds, sq.Eq{iampb.User_Id: ids})
	}

	if emails := req.GetEmails(); len(emails) != 0 {
		conds = append(conds, sq.Eq{iampb.User_Email: emails})
	}

	return conds
}

// List implements IUser.
func (u *Users) List(ctx context.Context,
	req *iampb.ListUsersRequest) (*iampb.ListUsersResponse, error) {

	conds := listConditions(req)

	builder := u.selectQuery(req.GetPage())
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

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
		total, err = u.total(ctx, conds)
		if err != nil {
			return nil, err
		}
	}

	return &iampb.ListUsersResponse{Users: users, Total: total}, nil
}

// total counts the rows matching the list conditions.
func (u *Users) total(
	ctx context.Context, conds []sq.Sqlizer,
) (int64, error) {
	builder := sq.Select("COUNT(*)").From(u.storage.Table())
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	var total int64
	if err := u.storage.Query(ctx, builder, &total); err != nil {
		return 0, err
	}

	return total, nil
}

// Create implements IUser.
func (u *Users) Create(ctx context.Context,
	user *iampb.User) (*iampb.User, error) {

	user.Id = rand.NewId(sztable.NewPrimaryKey(sztables.IAMUsers))
	query := szpostgres.InsertBuilder(u.storage, szpostgres.M{
		iampb.User_Id:           user.GetId(),
		iampb.User_Email:        user.GetEmail(),
		iampb.User_FullName:     user.GetFullName(),
		iampb.User_ControlPlane: user.GetControlPlane(),
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

// GetByEmail implements IUser. It returns errorx.ErrNotFound if no user
// owns the email.
func (u *Users) GetByEmail(
	ctx context.Context, email string) (*iampb.User, error) {

	builder := u.selectQuery(nil).Where(sq.Eq{iampb.User_Email: email})

	user, err := u.storage.CollectOneRow(ctx, builder, scanOne)
	if err != nil {
		if !errorx.IsNoRows(err) {
			return nil, err
		}

		return nil, errorx.ErrNotFound
	}

	return user, nil
}

// Update implements IUser.
func (u *Users) Update(ctx context.Context, user *iampb.User) error {

	query := szpostgres.UpdateBuilder(u.storage, user.GetId())

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
	return szpostgres.SelectBuilder(u.storage, page,
		iampb.User_Id,
		iampb.User_Email,
		iampb.User_FullName,
		iampb.User_ControlPlane,
		typepb.Metadata_CreatedAt,
		typepb.Metadata_UpdatedAt,
	)
}

func scan(rows szdbx.Rows) ([]*iampb.User, error) {
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

func scanOne(row szdbx.Row) (*iampb.User, error) {
	var (
		createdAt, updatedAt time.Time
		user                 iampb.User
	)

	err := row.Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
		&user.ControlPlane,
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
