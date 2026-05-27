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

package rulebased

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/postgres"
	"github.com/sentinez/core/storage/utils/table"
	"github.com/sentinez/core/tables"
	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/security/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// nolint
type IRuleBased interface {
	Create(ctx context.Context, model *securitypb.RuleBased) (*securitypb.RuleBased, error)
	Update(ctx context.Context, model *securitypb.RuleBased) error
	Get(ctx context.Context, id string) (*securitypb.RuleBased, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IRuleBased

	List(ctx context.Context, req *securitypb.ListRuleBasedsRequest) ([]*securitypb.RuleBased, int64, error)
}

func New(ctx context.Context, appConf *confpb.Config) (IRuleBased, error) {
	storage, err := postgres.New[securitypb.RuleBased](ctx, appConf,
		dbx.WithTable(tables.SecurityRuleBaseds),
		dbx.WithColumns(dbx.ColumnM{
			securitypb.RuleBased_Id:          postgres.String,
			securitypb.RuleBased_Name:        postgres.String,
			securitypb.RuleBased_Description: postgres.String,
			securitypb.RuleBased_Node:        postgres.JSONB,
			securitypb.RuleBased_Status:      postgres.Int4,
			securitypb.RuleBased_Priority:    postgres.Int4,
			securitypb.RuleBased_Action:      postgres.JSONB,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &RuleBased{
		storage: storage,
	}, nil
}

type RuleBased struct {
	storage dbx.Database[securitypb.RuleBased]
}

func (r *RuleBased) WithTX(tx *postgres.TxSession) IRuleBased {
	return &RuleBased{
		storage: postgres.WithTx(tx, r.storage),
	}
}

func buildListQuery(builder sq.SelectBuilder,
	req *securitypb.ListRuleBasedsRequest) sq.SelectBuilder {
	if len(req.GetIds()) > 0 {
		builder = builder.Where(sq.Eq{securitypb.RuleBased_Id: req.GetIds()})
	}
	return builder
}

func (r *RuleBased) List(ctx context.Context,
	req *securitypb.ListRuleBasedsRequest,
) ([]*securitypb.RuleBased, int64, error) {
	builder := postgres.SelectBuilder(r.storage, req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	zlog.Debug("[security.rulebaseds] query: ", query, " args: ", args)

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

func (r *RuleBased) Create(ctx context.Context,
	model *securitypb.RuleBased) (*securitypb.RuleBased, error) {
	model.Id = rand.NewID(table.NewPrimaryKey(tables.SecurityRuleBaseds))
	query := postgres.InsertBuilder(r.storage, postgres.M{
		securitypb.RuleBased_Id:          model.GetId(),
		securitypb.RuleBased_Name:        model.GetName(),
		securitypb.RuleBased_Description: model.GetDescription(),
		securitypb.RuleBased_Node:        model.GetNode(),
		securitypb.RuleBased_Status:      model.GetStatus(),
		securitypb.RuleBased_Priority:    model.GetPriority(),
		securitypb.RuleBased_Action:      model.GetAction(),
	})

	if _, err := r.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return model, nil
}

func (r *RuleBased) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *RuleBased) Get(
	ctx context.Context,
	id string,
) (*securitypb.RuleBased, error) {
	builder := r.selectQuery(nil).Where(sq.Eq{securitypb.RuleBased_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *RuleBased) Update(
	ctx context.Context,
	model *securitypb.RuleBased,
) error {
	query := postgres.UpdateBuilder(r.storage, model.GetId())

	if model.GetName() != "" {
		query = query.Set(securitypb.RuleBased_Name, model.GetName())
	}
	if model.GetDescription() != "" {
		query = query.Set(
			securitypb.RuleBased_Description,
			model.GetDescription(),
		)
	}
	if model.GetNode() != nil {
		query = query.Set(securitypb.RuleBased_Node, model.GetNode())
	}

	query = query.Set(securitypb.RuleBased_Status, model.GetStatus())

	_, err := r.storage.Exec(ctx, query)
	return err
}

func (r *RuleBased) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		securitypb.RuleBased_Id,
		securitypb.RuleBased_Name,
		securitypb.RuleBased_Description,
		securitypb.RuleBased_Node,
		securitypb.RuleBased_Status,
		securitypb.RuleBased_Priority,
		securitypb.RuleBased_Action,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*securitypb.RuleBased, error) {
	var models []*securitypb.RuleBased
	for rows.Next() {
		model, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func scanOne(row dbx.Row) (*securitypb.RuleBased, error) {
	var (
		createdAt, updatedAt time.Time
		model                securitypb.RuleBased
	)

	err := row.Scan(
		&model.Id,
		&model.Name,
		&model.Description,
		&model.Node,
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
