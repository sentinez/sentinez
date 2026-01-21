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

package accrepos

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
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ IAccount = (*Accounts)(nil)
)

// nolint
type IAccount interface {
	Create(ctx context.Context, account *AccountX) (*AccountX, error)
	Update(ctx context.Context, account *AccountX) error
	Get(ctx context.Context, id string) (*AccountX, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IAccount

	// extra methods

	GetByUsernameOrEmail(ctx context.Context, input string) (*AccountX, error)
	List(ctx context.Context, req *iam.ListAccountsRequest) (*iam.ListAccountsResponse, error)
	Total(ctx context.Context, req *iam.ListAccountsRequest) (int64, error)
}

func New(ctx context.Context, appConf *confpb.Config) (IAccount, error) {

	storage, err := postgres.New[AccountX](ctx, appConf,
		dbx.WithTable(tables.Accounts),
		dbx.WithColumn(iam.Account_Id, postgres.String),
		dbx.WithColumn(iam.Account_Email, postgres.String),
		dbx.WithColumn(iam.Account_Username, postgres.String),
		dbx.WithColumn(iam.Account_Password, postgres.String),
		dbx.WithColumn(iam.Account_Credentials, postgres.StringArr),
		dbx.WithColumn(iam.Account_UserId, postgres.String),
	)
	if err != nil {
		return nil, err
	}

	return &Accounts{
		storage: storage,
	}, nil
}

type Accounts struct {
	storage dbx.Database[AccountX]
}

// nolint:funlen
func buildListQuery(builder sq.SelectBuilder,
	req *iam.ListAccountsRequest) sq.SelectBuilder {

	if len(req.GetIds()) > 0 {
		builder = builder.Where(sq.Eq{iam.Account_Id: req.GetIds()})
	}

	if len(req.GetEmails()) > 0 {
		builder = builder.Where(sq.Eq{iam.Account_Email: req.GetEmails()})
	}

	if len(req.GetUserIds()) > 0 {
		builder = builder.Where(sq.Eq{iam.Account_UserId: req.GetUserIds()})
	}

	if len(req.GetUsernames()) > 0 {
		builder = builder.Where(sq.Eq{iam.Account_Username: req.GetUsernames()})
	}

	return builder
}

func (acc *Accounts) WithTX(tx *postgres.TxSession) IAccount {
	return &Accounts{
		storage: postgres.WithTx(tx, acc.storage),
	}
}

// nolint:funlen
func (acc *Accounts) List(ctx context.Context,
	req *iam.ListAccountsRequest) (*iam.ListAccountsResponse, error) {

	builder := acc.selectQuery(req.GetPage())
	builder = buildListQuery(builder, req)

	accounts, err := acc.storage.CollectRows(ctx, builder, scan)
	if err != nil {
		return nil, err
	}

	var resp iam.ListAccountsResponse
	for _, account := range accounts {
		resp.Accounts = append(resp.Accounts, &iam.AccountResponse{
			Id:       account.GetId(),
			Username: account.GetUsername(),
			Email:    account.GetEmail(),
			UserId:   account.GetUserId(),
		})
	}

	if req.GetPage().GetTotal() {
		resp.Total, err = acc.Total(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	return &resp, nil
}

func (acc *Accounts) Total(ctx context.Context,
	req *iam.ListAccountsRequest) (int64, error) {

	builder := sq.Select("COUNT(*) AS count").From(acc.storage.Table())
	builder = buildListQuery(builder, req)

	var count int64
	err := acc.storage.Query(ctx, builder, &count)

	return count, err
}

// GetByUsernameOrEmail implements IAccount.
func (acc *Accounts) GetByUsernameOrEmail(ctx context.Context,
	input string) (*AccountX, error) {

	builder := acc.selectQuery(nil).Where(sq.Or{
		sq.Eq{iam.Account_Username: input},
		sq.Eq{iam.Account_Email: input},
	})

	resp, err := acc.storage.CollectOneRow(ctx, builder, scanOne)
	if err != nil {
		return &AccountX{}, err
	}

	return resp, nil
}

// Create implements IAccount.
func (acc *Accounts) Create(ctx context.Context,
	account *AccountX) (*AccountX, error) {

	account.Id = ids.NewID(table.NewPrimaryKey(tables.Accounts))

	if account.GetUsername() == "" {
		account.Username = account.GetEmail()
	}

	query := postgres.InsertBuilder(acc.storage,
		[]string{
			iam.Account_Id,
			iam.Account_Credentials,
			iam.Account_Username,
			iam.Account_Email,
			iam.Account_UserId,
			iam.Account_Password,
		},
		[]any{
			account.GetId(),
			account.GetCredentials(),
			account.GetUsername(),
			account.GetEmail(),
			account.GetUserId(),
			account.GetPassword(),
		},
	)

	_, err := acc.storage.Insert(ctx, query)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Delete implements IAccount.
func (acc *Accounts) Delete(ctx context.Context, id string) error {
	return acc.storage.Delete(ctx, id)
}

// Get implements IAccount.
func (acc *Accounts) Get(ctx context.Context, id string) (*AccountX, error) {
	builder := acc.selectQuery(nil).Where(sq.Eq{iam.Account_Id: id})
	return acc.storage.Select(ctx, builder, scanOne)
}

// Update implements IAccount.
func (acc *Accounts) Update(ctx context.Context, account *AccountX) error {

	query := postgres.UpdateBuilder(acc.storage, account.GetId())

	if account.GetEmail() != "" {
		query = query.Set(iam.Account_Email, account.GetEmail())
	}

	if account.GetUsername() != "" {
		query = query.Set(iam.Account_Username, account.GetUsername())
	}

	if account.GetPassword() != "" {
		query = query.Set(iam.Account_Password, account.GetPassword())
	}

	if account.GetUserId() != "" {
		query = query.Set(iam.Account_UserId, account.GetUserId())
	}

	if len(account.GetCredentials()) != 0 {
		query = query.Set(iam.Account_Credentials, account.GetCredentials())
	}

	_, err := acc.storage.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (acc *Accounts) selectQuery(page *common.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(acc.storage, page,
		iam.Account_Id,
		iam.Account_Credentials,
		iam.Account_Username,
		iam.Account_Email,
		iam.Account_UserId,
		iam.Account_Password,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*AccountX, error) {
	var results []*AccountX

	for rows.Next() {
		var createdAt, updatedAt time.Time
		account := AccountX{Account: &iam.Account{}}
		err := rows.Scan(
			&account.Id,
			&account.Credentials,
			&account.Username,
			&account.Email,
			&account.UserId,
			&account.Password,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		account.Metadata = &modelpb.Metadata{
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: timestamppb.New(updatedAt),
		}

		results = append(results, &account)
	}

	return results, rows.Err()
}

func scanOne(row dbx.Row) (*AccountX, error) {

	var createdAt, updatedAt time.Time
	account := AccountX{Account: &iam.Account{}}
	err := row.Scan(
		&account.Id,
		&account.Credentials,
		&account.Username,
		&account.Email,
		&account.UserId,
		&account.Password,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	account.Metadata = &modelpb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &account, nil
}
