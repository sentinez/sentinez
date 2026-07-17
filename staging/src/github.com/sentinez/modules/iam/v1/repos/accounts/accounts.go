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
	"github.com/sentinez/core/common/tables"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/postgres"
	"github.com/sentinez/core/storage/utils/table"
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
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
	List(ctx context.Context, req *iampb.ListAccountsRequest) (*iampb.ListAccountsResponse, error)
}

func New(ctx context.Context, appConf *settingpb.Config) (IAccount, error) {

	storage, err := postgres.New[AccountX](ctx, appConf,
		dbx.WithTable(tables.IAMAccounts),
		dbx.WithColumns(dbx.ColumnM{
			iampb.Account_Id:             postgres.String,
			iampb.Account_Email:          postgres.String,
			iampb.Account_Username:       postgres.String,
			iampb.Account_PasswordHash:   postgres.String,
			iampb.Account_Credentials:    postgres.StringArr,
			iampb.Account_UserId:         postgres.String,
			iampb.Account_Provider:       postgres.String,
			iampb.Account_ProviderUserId: postgres.String,
		}),
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
	req *iampb.ListAccountsRequest) sq.SelectBuilder {

	if len(req.GetIds()) > 0 {
		builder = builder.Where(sq.Eq{iampb.Account_Id: req.GetIds()})
	}

	if len(req.GetEmails()) > 0 {
		builder = builder.Where(sq.Eq{iampb.Account_Email: req.GetEmails()})
	}

	if len(req.GetUserIds()) > 0 {
		builder = builder.Where(sq.Eq{iampb.Account_UserId: req.GetUserIds()})
	}

	if len(req.GetUsernames()) > 0 {
		builder = builder.Where(
			sq.Eq{iampb.Account_Username: req.GetUsernames()})
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
	req *iampb.ListAccountsRequest) (*iampb.ListAccountsResponse, error) {

	builder := acc.selectQuery(req.GetPage())
	builder = buildListQuery(builder, req)

	accounts, err := acc.storage.CollectRows(ctx, builder, scan)
	if err != nil {
		return nil, err
	}

	var resp iampb.ListAccountsResponse
	for _, account := range accounts {
		resp.Accounts = append(resp.Accounts, &iampb.AccountResponse{
			Id:       account.GetId(),
			Username: account.GetUsername(),
			Email:    account.GetEmail(),
			UserId:   account.GetUserId(),
		})
	}

	if req.GetPage().GetTotal() {
		resp.Total, err = acc.storage.Total(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &resp, nil
}

// GetByUsernameOrEmail implements IAccount.
func (acc *Accounts) GetByUsernameOrEmail(ctx context.Context,
	input string) (*AccountX, error) {

	builder := acc.selectQuery(nil).Where(sq.Or{
		sq.Eq{iampb.Account_Username: input},
		sq.Eq{iampb.Account_Email: input},
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

	account.Id = rand.NewID(table.NewPrimaryKey(tables.IAMAccounts))

	if account.GetUsername() == "" {
		account.Username = account.GetEmail()
	}

	query := postgres.InsertBuilder(acc.storage, postgres.M{
		iampb.Account_Id:             account.GetId(),
		iampb.Account_Credentials:    account.GetCredentials(),
		iampb.Account_Username:       account.GetUsername(),
		iampb.Account_Email:          account.GetEmail(),
		iampb.Account_UserId:         account.GetUserId(),
		iampb.Account_PasswordHash:   account.GetPasswordHash(),
		iampb.Account_Provider:       account.GetProvider(),
		iampb.Account_ProviderUserId: account.GetProviderUserId(),
	})

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
	builder := acc.selectQuery(nil).Where(sq.Eq{iampb.Account_Id: id})
	return acc.storage.Select(ctx, builder, scanOne)
}

// Update implements IAccount.
// nolint:funlen
func (acc *Accounts) Update(ctx context.Context, account *AccountX) error {

	query := postgres.UpdateBuilder(acc.storage, account.GetId())

	if account.GetEmail() != "" {
		query = query.Set(iampb.Account_Email, account.GetEmail())
	}

	if account.GetUsername() != "" {
		query = query.Set(iampb.Account_Username, account.GetUsername())
	}

	if account.GetPasswordHash() != "" {
		query = query.Set(iampb.Account_PasswordHash, account.GetPasswordHash())
	}

	if account.GetUserId() != "" {
		query = query.Set(iampb.Account_UserId, account.GetUserId())
	}

	if len(account.GetCredentials()) != 0 {
		query = query.Set(iampb.Account_Credentials, account.GetCredentials())
	}

	if account.GetProvider() != "" {
		query = query.Set(iampb.Account_Provider, account.GetProvider())
	}

	if account.GetProviderUserId() != "" {
		query = query.Set(
			iampb.Account_ProviderUserId, account.GetProviderUserId())
	}

	_, err := acc.storage.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (acc *Accounts) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(acc.storage, page,
		iampb.Account_Id,
		iampb.Account_Credentials,
		iampb.Account_Username,
		iampb.Account_Email,
		iampb.Account_UserId,
		iampb.Account_PasswordHash,
		iampb.Account_Provider,
		iampb.Account_ProviderUserId,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*AccountX, error) {
	var accounts []*AccountX

	for rows.Next() {
		account, err := scanOne(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

func scanOne(row dbx.Row) (*AccountX, error) {
	var createdAt, updatedAt time.Time
	account := AccountX{Account: &iampb.Account{}}

	err := row.Scan(
		&account.Id,
		&account.Credentials,
		&account.Username,
		&account.Email,
		&account.UserId,
		&account.PasswordHash,
		&account.Provider,
		&account.ProviderUserId,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		zlog.Debugf("postgres: in tx=false query: %s", row)
		return nil, err
	}

	account.Metadata = &typepb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &account, nil
}
