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

package rulegroups

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

// nolint
type IRuleGroup interface {
	Create(ctx context.Context, model *securitypb.RuleGroup) (*securitypb.RuleGroup, error)
	Update(ctx context.Context, model *securitypb.RuleGroup) error
	Get(ctx context.Context, id string) (*securitypb.RuleGroup, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IRuleGroup

	List(ctx context.Context, req *securitypb.ListRuleGroupsRequest) ([]*securitypb.RuleGroup, int64, error)
}

func New(ctx context.Context, appConf *confpb.Config) (IRuleGroup, error) {
	storage, err := postgres.New[securitypb.RuleGroup](ctx, appConf,
		dbx.WithTable(tables.SecurityRuleGroups),
		dbx.WithColumns(dbx.ColumnM{
			securitypb.RuleGroup_Id:          postgres.String,
			securitypb.RuleGroup_Name:        postgres.String,
			securitypb.RuleGroup_Description: postgres.String,
			securitypb.RuleGroup_Node:        postgres.JSONB,
			securitypb.RuleGroup_Status:      postgres.Int4,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &RuleGroups{
		storage: storage,
	}, nil
}

type RuleGroups struct {
	storage dbx.Database[securitypb.RuleGroup]
}

func (e *RuleGroups) WithTX(tx *postgres.TxSession) IRuleGroup {
	return &RuleGroups{
		storage: postgres.WithTx(tx, e.storage),
	}
}

func buildListQuery(builder sq.SelectBuilder,
	req *securitypb.ListRuleGroupsRequest) sq.SelectBuilder {
	if len(req.GetIds()) > 0 {
		builder = builder.Where(sq.Eq{securitypb.RuleGroup_Id: req.GetIds()})
	}
	return builder
}

func (e *RuleGroups) List(ctx context.Context,
	req *securitypb.ListRuleGroupsRequest,
) ([]*securitypb.RuleGroup, int64, error) {
	builder := postgres.SelectBuilder(e.storage, req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	zlog.Debug("[security.rulegroups] query: ", query, " args: ", args)

	models, err := e.storage.CollectRows(ctx, builder, scan)
	if err != nil {
		return nil, 0, err
	}

	if req.GetPage().GetTotal() {
		total, err = e.storage.Total(ctx)
		if err != nil {
			return nil, 0, err
		}
	}

	return models, total, nil
}

func (e *RuleGroups) Create(ctx context.Context,
	model *securitypb.RuleGroup) (*securitypb.RuleGroup, error) {
	model.Id = ids.NewID(table.NewPrimaryKey(tables.SecurityRuleGroups))
	query := postgres.InsertBuilder(e.storage, postgres.M{
		securitypb.RuleGroup_Id:          model.GetId(),
		securitypb.RuleGroup_Name:        model.GetName(),
		securitypb.RuleGroup_Description: model.GetDescription(),
		securitypb.RuleGroup_Node:        model.GetNode(),
		securitypb.RuleGroup_Status:      model.GetStatus(),
	})

	if _, err := e.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return model, nil
}

func (e *RuleGroups) Delete(ctx context.Context, id string) error {
	return e.storage.Delete(ctx, id)
}

func (e *RuleGroups) Get(
	ctx context.Context,
	id string,
) (*securitypb.RuleGroup, error) {
	builder := e.selectQuery(nil).Where(sq.Eq{securitypb.RuleGroup_Id: id})
	return e.storage.Select(ctx, builder, scanOne)
}

func (e *RuleGroups) Update(
	ctx context.Context,
	model *securitypb.RuleGroup,
) error {
	query := postgres.UpdateBuilder(e.storage, model.GetId())

	if model.GetName() != "" {
		query = query.Set(securitypb.RuleGroup_Name, model.GetName())
	}
	if model.GetDescription() != "" {
		query = query.Set(
			securitypb.RuleGroup_Description,
			model.GetDescription(),
		)
	}
	if model.GetNode() != "" {
		query = query.Set(securitypb.RuleGroup_Node, model.GetNode())
	}

	query = query.Set(securitypb.RuleGroup_Status, model.GetStatus())

	_, err := e.storage.Exec(ctx, query)
	return err
}

func (e *RuleGroups) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(e.storage, page,
		securitypb.RuleGroup_Id,
		securitypb.RuleGroup_Name,
		securitypb.RuleGroup_Description,
		securitypb.RuleGroup_Node,
		securitypb.RuleGroup_Status,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*securitypb.RuleGroup, error) {
	var models []*securitypb.RuleGroup
	for rows.Next() {
		model, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func scanOne(row dbx.Row) (*securitypb.RuleGroup, error) {
	var (
		createdAt, updatedAt time.Time
		model                securitypb.RuleGroup
	)

	err := row.Scan(
		&model.Id,
		&model.Name,
		&model.Description,
		&model.Node,
		&model.Status,
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
