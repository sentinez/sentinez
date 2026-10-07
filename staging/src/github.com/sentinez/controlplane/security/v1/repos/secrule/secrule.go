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

package secrule

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/core/common/tables"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/postgres"
	"github.com/sentinez/core/storage/utils/table"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// nolint
type ISecRule interface {
	Create(ctx context.Context, model *securitypb.SecRule) (*securitypb.SecRule, error)
	Update(ctx context.Context, model *securitypb.SecRule) error
	Get(ctx context.Context, id string) (*securitypb.SecRule, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) ISecRule

	List(ctx context.Context, req *securitypb.ListSecRulesRequest) ([]*securitypb.SecRule, int64, error)
}

func New(ctx context.Context, appConf *settingpb.Config) (ISecRule, error) {
	storage, err := postgres.New[securitypb.SecRule](ctx, appConf,
		dbx.WithTable(tables.SecuritySecRules),
		dbx.WithColumns(dbx.ColumnM{
			securitypb.SecRule_Id:          postgres.String,
			securitypb.SecRule_Name:        postgres.String,
			securitypb.SecRule_Description: postgres.String,
			securitypb.SecRule_Expr:        postgres.JSONB,
			securitypb.SecRule_Status:      postgres.Int4,
			securitypb.SecRule_Priority:    postgres.Int4,
			securitypb.SecRule_Action:      postgres.JSONB,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &SecRule{
		storage: storage,
	}, nil
}

type SecRule struct {
	storage dbx.Database[securitypb.SecRule]
}

func (r *SecRule) WithTX(tx *postgres.TxSession) ISecRule {
	return &SecRule{
		storage: postgres.WithTx(tx, r.storage),
	}
}

func buildListQuery(builder sq.SelectBuilder,
	req *securitypb.ListSecRulesRequest) sq.SelectBuilder {
	if len(req.GetIds()) > 0 {
		builder = builder.Where(sq.Eq{securitypb.SecRule_Id: req.GetIds()})
	}
	return builder
}

func (r *SecRule) List(ctx context.Context,
	req *securitypb.ListSecRulesRequest,
) ([]*securitypb.SecRule, int64, error) {
	builder := postgres.SelectBuilder(r.storage, req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	zlog.Debug("[security.secrules] query: ", query, " args: ", args)

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

func (r *SecRule) Create(ctx context.Context,
	model *securitypb.SecRule) (*securitypb.SecRule, error) {
	model.Id = rand.NewID(table.NewPrimaryKey(tables.SecuritySecRules))
	query := postgres.InsertBuilder(r.storage, postgres.M{
		securitypb.SecRule_Id:          model.GetId(),
		securitypb.SecRule_Name:        model.GetName(),
		securitypb.SecRule_Description: model.GetDescription(),
		securitypb.SecRule_Expr:        model.GetExpr(),
		securitypb.SecRule_Status:      model.GetStatus(),
		securitypb.SecRule_Priority:    model.GetPriority(),
		securitypb.SecRule_Action:      model.GetAction(),
	})

	if _, err := r.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return model, nil
}

func (r *SecRule) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *SecRule) Get(
	ctx context.Context,
	id string,
) (*securitypb.SecRule, error) {
	builder := r.selectQuery(nil).Where(sq.Eq{securitypb.SecRule_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *SecRule) Update(
	ctx context.Context,
	model *securitypb.SecRule,
) error {
	query := postgres.UpdateBuilder(r.storage, model.GetId())

	if model.GetName() != "" {
		query = query.Set(securitypb.SecRule_Name, model.GetName())
	}
	if model.GetDescription() != "" {
		query = query.Set(
			securitypb.SecRule_Description,
			model.GetDescription(),
		)
	}
	if model.GetExpr() != nil {
		query = query.Set(securitypb.SecRule_Expr, model.GetExpr())
	}

	query = query.Set(securitypb.SecRule_Status, model.GetStatus())

	_, err := r.storage.Exec(ctx, query)
	return err
}

func (r *SecRule) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		securitypb.SecRule_Id,
		securitypb.SecRule_Name,
		securitypb.SecRule_Description,
		securitypb.SecRule_Expr,
		securitypb.SecRule_Status,
		securitypb.SecRule_Priority,
		securitypb.SecRule_Action,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*securitypb.SecRule, error) {
	var models []*securitypb.SecRule
	for rows.Next() {
		model, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func scanOne(row dbx.Row) (*securitypb.SecRule, error) {
	var (
		createdAt, updatedAt time.Time
		model                securitypb.SecRule
	)

	err := row.Scan(
		&model.Id,
		&model.Name,
		&model.Description,
		&model.Expr,
		&model.Status,
		&model.Priority,
		&model.Action,
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
