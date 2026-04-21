// Copyright 2025 Sentinéz Labs.
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

package rules

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/security/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/shared/tables"
	"github.com/sentinez/sentinez/pkg/storage/dbx"
	"github.com/sentinez/sentinez/pkg/storage/dbx/postgres"
	"github.com/sentinez/sentinez/pkg/storage/utils/table"
	"github.com/sentinez/shared/ids"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ IRule = (*Rules)(nil)
)

// nolint
type IRule interface {
	Create(ctx context.Context, model *securitypb.Rule) (*securitypb.Rule, error)
	Update(ctx context.Context, model *securitypb.Rule) error
	Get(ctx context.Context, id string) (*securitypb.Rule, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IRule

	List(ctx context.Context, req *securitypb.ListRulesRequest) ([]*securitypb.Rule, int64, error)
}

func New(ctx context.Context, appConf *confpb.Config) (IRule, error) {
	storage, err := postgres.New[securitypb.Rule](ctx, appConf,
		dbx.WithTable(tables.SecurityRules),
		dbx.WithColumns(dbx.ColumnM{
			securitypb.Rule_Id:          postgres.String,
			securitypb.Rule_Name:        postgres.String,
			securitypb.Rule_Description: postgres.String,
			securitypb.Rule_Rule:        postgres.String,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Rules{
		storage: storage,
	}, nil
}

type Rules struct {
	storage dbx.Database[securitypb.Rule]
}

func (r *Rules) WithTX(tx *postgres.TxSession) IRule {
	return &Rules{
		storage: postgres.WithTx(tx, r.storage),
	}
}

func buildListQuery(builder sq.SelectBuilder,
	req *securitypb.ListRulesRequest) sq.SelectBuilder {
	if len(req.GetIds()) > 0 {
		builder = builder.Where(sq.Eq{securitypb.Rule_Id: req.GetIds()})
	}
	return builder
}

func (r *Rules) List(ctx context.Context,
	req *securitypb.ListRulesRequest) ([]*securitypb.Rule, int64, error) {
	builder := postgres.SelectBuilder(r.storage, req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	zlog.Debug("[security.rules] query: ", query, " args: ", args)

	models, err := r.storage.CollectRows(ctx, builder, scan)
	if err != nil {
		return nil, 0, err
	}

	if req.GetPage().GetTotal() {
		total, err = r.storage.Total(ctx)
		if err != nil {
			return nil, 0, err
		}
	}

	return models, total, nil
}

func (r *Rules) Create(ctx context.Context,
	model *securitypb.Rule) (*securitypb.Rule, error) {
	model.Id = ids.NewID(table.NewPrimaryKey(tables.SecurityRules))
	query := postgres.InsertBuilder(r.storage, postgres.M{
		securitypb.Rule_Id:          model.GetId(),
		securitypb.Rule_Name:        model.GetName(),
		securitypb.Rule_Description: model.GetDescription(),
		securitypb.Rule_Rule:        model.GetRule(),
	})

	if _, err := r.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return model, nil
}

func (r *Rules) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *Rules) Get(ctx context.Context, id string) (*securitypb.Rule, error) {
	builder := r.selectQuery(nil).Where(sq.Eq{securitypb.Rule_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *Rules) Update(ctx context.Context, model *securitypb.Rule) error {
	query := postgres.UpdateBuilder(r.storage, model.GetId())

	if model.GetName() != "" {
		query = query.Set(securitypb.Rule_Name, model.GetName())
	}
	if model.GetDescription() != "" {
		query = query.Set(securitypb.Rule_Description, model.GetDescription())
	}
	if model.GetRule() != "" {
		query = query.Set(securitypb.Rule_Rule, model.GetRule())
	}

	_, err := r.storage.Exec(ctx, query)
	return err
}

func (r *Rules) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		securitypb.Rule_Id,
		securitypb.Rule_Name,
		securitypb.Rule_Description,
		securitypb.Rule_Rule,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*securitypb.Rule, error) {
	var models []*securitypb.Rule
	for rows.Next() {
		model, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func scanOne(row dbx.Row) (*securitypb.Rule, error) {
	var (
		createdAt, updatedAt time.Time
		model                securitypb.Rule
	)

	err := row.Scan(
		&model.Id,
		&model.Name,
		&model.Description,
		&model.Rule,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	model.Metadata = &typepb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &model, nil
}
