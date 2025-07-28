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

package accountrepo

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/pkg/auto/queries/accounts"
	"github.com/sentinez/sentinez/pkg/common/uuid"
	"github.com/sentinez/sentinez/pkg/infra/database"
	"github.com/sentinez/sentinez/pkg/infra/database/postgresz"
	"github.com/sentinez/sentinez/pkg/std/table"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	_ database.Repository[*iam.Accounts, string] = (*Accounts)(nil)
	_ IAccount                                   = (*Accounts)(nil)
)

type IAccount interface {
	Create(ctx context.Context, account *iam.Accounts) (*iam.Accounts, error)
	Update(ctx context.Context, account *iam.Accounts) (*iam.Accounts, error)
	Get(ctx context.Context, id string) (*iam.Accounts, error)
	GetMany(ctx context.Context, page *common.Pages) ([]*iam.Accounts, error)
	Delete(ctx context.Context, id string) error

	// extra methods

	GetByUsernameOrEmail(ctx context.Context,
		input string) (*iam.Accounts, error)

	List(ctx context.Context,
		req *iam.ListAccountsRequest) (*iam.ListAccountsResponse, error)
}

func New(pool *pgxpool.Pool) (IAccount, error) {
	tableName := table.Table(table.Account)

	storage, err := postgresz.New[*iam.Accounts](pool, tableName,
		postgresz.WithIndex("username", "user_id", "email"))
	if err != nil {
		return nil, err
	}

	return &Accounts{
		query:     accounts.New(pool),
		storage:   storage,
		tableName: tableName,
	}, nil
}

type Accounts struct {
	tableName string
	query     *accounts.Queries
	storage   database.Database[*iam.Accounts]
}

// nolint:funlen
func (acc *Accounts) List(ctx context.Context,
	req *iam.ListAccountsRequest) (*iam.ListAccountsResponse, error) {

	builder := database.Paging(acc.storage.SelectBuilder(), req.GetPage())

	if len(req.GetIds()) > 0 {
		builder = builder.Where(
			squirrel.Eq{postgresz.Primary(iam.AccountsFieldId): req.GetIds()})
	}

	if len(req.GetEmails()) > 0 {
		builder = builder.Where(squirrel.Eq{
			postgresz.Field(iam.AccountsFieldEmail): req.GetEmails(),
		})
	}

	if len(req.GetUserIds()) > 0 {
		builder = builder.Where(squirrel.Eq{
			postgresz.Field(iam.AccountsFieldUserId): req.GetUserIds(),
		})
	}

	if len(req.GetUsernames()) > 0 {
		builder = builder.Where(squirrel.Eq{
			postgresz.Field(iam.AccountsFieldUsername): req.GetUsernames(),
		})
	}

	accounts, err := acc.storage.CollectRows(
		ctx, builder, postgresz.Scans[*iam.Accounts])
	if err != nil {
		return nil, err
	}

	var resp iam.ListAccountsResponse
	for _, account := range accounts {
		resp.Accounts = append(resp.Accounts, &iam.AccountLite{
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
	input string) (*iam.Accounts, error) {

	account, err := acc.query.GetByUsernameOrEmail(ctx, []byte(input))
	if err != nil {
		return nil, err
	}

	var result iam.Accounts
	if err := protojson.Unmarshal(account.Data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Create implements IAccount.
func (acc *Accounts) Create(ctx context.Context,
	account *iam.Accounts) (*iam.Accounts, error) {

	account.Id = uuid.Generate(acc.tableName)

	data, err := protojson.Marshal(account)
	if err != nil {
		return nil, err
	}

	err = acc.query.Insert(ctx,
		accounts.InsertParams{ID: account.GetId(), Column2: data})
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Delete implements IAccount.
func (acc *Accounts) Delete(ctx context.Context, id string) error {
	return acc.query.Delete(ctx, id)
}

// Get implements IAccount.
func (acc *Accounts) Get(ctx context.Context,
	id string) (*iam.Accounts, error) {

	account, err := acc.query.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var result iam.Accounts
	if err := protojson.Unmarshal(account.Data, &result); err != nil {
		return nil, err
	}

	return &result, err
}

// GetMany implements IAccount.
func (acc *Accounts) GetMany(ctx context.Context,
	page *common.Pages) ([]*iam.Accounts, error) {

	builder := database.Paging(acc.storage.SelectBuilder(), page)

	resp, err := acc.storage.CollectRows(
		ctx, builder, postgresz.Scans[*iam.Accounts])
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Update implements IAccount.
func (acc *Accounts) Update(
	ctx context.Context, account *iam.Accounts) (*iam.Accounts, error) {

	data, err := protojson.Marshal(account)
	if err != nil {
		return nil, err
	}

	err = acc.query.Update(ctx,
		accounts.UpdateParams{ID: account.GetId(), Column2: data})
	if err != nil {
		return nil, err
	}

	return account, nil
}
