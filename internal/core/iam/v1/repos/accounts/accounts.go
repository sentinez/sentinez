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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	commonpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	modelpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/model/v1"
	"github.com/sentinez/sentinez/internal/common/tables"
	"github.com/sentinez/sentinez/pkg/common/uuid"
	"github.com/sentinez/sentinez/pkg/storage/database"
	"github.com/sentinez/sentinez/pkg/storage/database/postgres"
	"github.com/sentinez/sentinez/pkg/table"

	sq "github.com/Masterminds/squirrel"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ IAccount = (*Accounts)(nil)
)

type IAccount interface {
	Create(ctx context.Context, account *iam.Accounts) (*iam.Accounts, error)
	Update(ctx context.Context, account *iam.Accounts) (*iam.Accounts, error)
	Get(ctx context.Context, id string) (*iam.Accounts, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IAccount

	// extra methods

	GetByUsernameOrEmail(ctx context.Context,
		input string) (*iam.Accounts, error)

	List(ctx context.Context,
		req *iam.ListAccountsRequest) (*iam.ListAccountsResponse, error)

	Total(ctx context.Context, req *iam.ListAccountsRequest) (int64, error)
}

func New(appConf *commonpb.AppConfig) (IAccount, error) {

	storage, err := postgres.New[*iam.Accounts](appConf, tables.Accounts,
		postgres.WithIndex("username", "user_id", "email"))
	if err != nil {
		return nil, err
	}

	return &Accounts{
		storage: storage,
	}, nil
}

type Accounts struct {
	storage database.Database[*iam.Accounts]
}

// nolint:funlen
func buildListQuery(builder sq.SelectBuilder,
	req *iam.ListAccountsRequest) sq.SelectBuilder {

	if len(req.GetIds()) > 0 {
		builder = builder.Where(
			sq.Eq{postgres.Primary(iam.AccountsFieldId): req.GetIds()})
	}

	if len(req.GetEmails()) > 0 {
		builder = builder.Where(sq.Eq{
			postgres.Field(iam.AccountsFieldEmail): req.GetEmails(),
		})
	}

	if len(req.GetUserIds()) > 0 {
		builder = builder.Where(sq.Eq{
			postgres.Field(iam.AccountsFieldUserId): req.GetUserIds(),
		})
	}

	if len(req.GetUsernames()) > 0 {
		builder = builder.Where(sq.Eq{
			postgres.Field(iam.AccountsFieldUsername): req.GetUsernames(),
		})
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

	builder := sq.Select(database.SchemalessFieldData).
		From(acc.storage.Table())
	builder = postgres.Paging(builder, req.GetPage())
	builder = buildListQuery(builder, req)

	accounts, err := acc.storage.CollectRows(
		ctx, builder, postgres.Scans[*iam.Accounts])
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
			Metadata: account.GetMetadata(),
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
	input string) (*iam.Accounts, error) {

	builder := sq.Select(database.SchemalessFieldData).
		From(acc.storage.Table()).
		Where(sq.Or{
			sq.Eq{postgres.Field(iam.AccountsFieldUsername): input},
			sq.Eq{postgres.Field(iam.AccountsFieldEmail): input},
		})

	return acc.storage.
		CollectOneRow(ctx, builder, postgres.Scan[*iam.Accounts])
}

// Create implements IAccount.
func (acc *Accounts) Create(ctx context.Context,
	account *iam.Accounts) (*iam.Accounts, error) {

	now := timestamppb.Now()
	account.Id = uuid.NewID(table.NewPrimaryKey(tables.Accounts))
	account.Metadata = &modelpb.Metadata{
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       account.GetUsername(),
		UpdatedBy:       account.GetUsername(),
		ResourceOwnerId: account.GetUserId(),
	}

	if err := acc.storage.Set(ctx, account.GetId(), account); err != nil {
		return nil, err
	}

	return account, nil
}

// Delete implements IAccount.
func (acc *Accounts) Delete(ctx context.Context, id string) error {
	return acc.storage.Delete(ctx, id)
}

// Get implements IAccount.
func (acc *Accounts) Get(ctx context.Context,
	id string) (*iam.Accounts, error) {

	return acc.storage.Get(ctx, id)
}

// Update implements IAccount.
func (acc *Accounts) Update(
	ctx context.Context, account *iam.Accounts) (*iam.Accounts, error) {

	account.Metadata.UpdatedAt = timestamppb.Now()
	if err := acc.storage.Set(ctx, account.GetId(), account); err != nil {
		return nil, err
	}

	return account, nil
}
