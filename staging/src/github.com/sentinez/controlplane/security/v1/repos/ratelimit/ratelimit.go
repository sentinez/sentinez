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

package ratelimit

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
type IRateLimit interface {
	Create(ctx context.Context, model *securitypb.RateLimit) (*securitypb.RateLimit, error)
	Update(ctx context.Context, model *securitypb.RateLimit) error
	Get(ctx context.Context, id string) (*securitypb.RateLimit, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IRateLimit

	List(ctx context.Context, req *securitypb.ListRateLimitsRequest) ([]*securitypb.RateLimit, int64, error)
}

var _ IRateLimit = (*RateLimit)(nil)

// New creates a RateLimit repository backed by Postgres.
func New(ctx context.Context, appConf *settingpb.Config) (IRateLimit, error) {
	storage, err := postgres.New[securitypb.RateLimit](ctx, appConf,
		dbx.WithTable(tables.SecurityRateLimits),
		dbx.WithColumns(dbx.ColumnM{
			securitypb.RateLimit_Id:          postgres.String,
			securitypb.RateLimit_Name:        postgres.String,
			securitypb.RateLimit_Description: postgres.String,
			securitypb.RateLimit_Expr:        postgres.JSONB,
			securitypb.RateLimit_Status:      postgres.Int4,
			securitypb.RateLimit_Priority:    postgres.Int4,
			securitypb.RateLimit_Action:      postgres.String,
			securitypb.RateLimit_TimeWindow:  postgres.String,
			securitypb.RateLimit_MaxRequests: postgres.Int8,
			securitypb.RateLimit_Timeout:     postgres.String,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &RateLimit{
		storage: storage,
	}, nil
}

// RateLimit is the Postgres repository for rate limit rules.
type RateLimit struct {
	storage dbx.Database[securitypb.RateLimit]
}

func (r *RateLimit) WithTX(tx *postgres.TxSession) IRateLimit {
	return &RateLimit{
		storage: postgres.WithTx(tx, r.storage),
	}
}

func buildListQuery(builder sq.SelectBuilder,
	req *securitypb.ListRateLimitsRequest) sq.SelectBuilder {
	if len(req.GetIds()) > 0 {
		builder = builder.Where(sq.Eq{securitypb.RateLimit_Id: req.GetIds()})
	}
	return builder
}

func (r *RateLimit) List(ctx context.Context,
	req *securitypb.ListRateLimitsRequest,
) ([]*securitypb.RateLimit, int64, error) {
	builder := r.selectQuery(req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	zlog.Debug("[security.ratelimits] query: ", query, " args: ", args)

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

func (r *RateLimit) Create(ctx context.Context,
	model *securitypb.RateLimit) (*securitypb.RateLimit, error) {
	model.Id = rand.NewID(table.NewPrimaryKey(tables.SecurityRateLimits))
	query := postgres.InsertBuilder(r.storage, postgres.M{
		securitypb.RateLimit_Id:          model.GetId(),
		securitypb.RateLimit_Name:        model.GetName(),
		securitypb.RateLimit_Description: model.GetDescription(),
		securitypb.RateLimit_Expr:        model.GetExpr(),
		securitypb.RateLimit_Status:      model.GetStatus(),
		securitypb.RateLimit_Priority:    model.GetPriority(),
		securitypb.RateLimit_Action:      model.GetAction(),
		securitypb.RateLimit_TimeWindow:  model.GetTimeWindow(),
		securitypb.RateLimit_MaxRequests: model.GetMaxRequests(),
		securitypb.RateLimit_Timeout:     model.GetTimeout(),
	})

	if _, err := r.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return model, nil
}

func (r *RateLimit) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *RateLimit) Get(
	ctx context.Context,
	id string,
) (*securitypb.RateLimit, error) {
	builder := r.selectQuery(nil).
		Where(sq.Eq{securitypb.RateLimit_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *RateLimit) Update(
	ctx context.Context,
	model *securitypb.RateLimit,
) error {
	query := postgres.UpdateBuilder(r.storage, model.GetId()).
		SetMap(sq.Eq{
			securitypb.RateLimit_Name:        model.GetName(),
			securitypb.RateLimit_Description: model.GetDescription(),
			securitypb.RateLimit_Expr:        model.GetExpr(),
			securitypb.RateLimit_Status:      model.GetStatus(),
			securitypb.RateLimit_Priority:    model.GetPriority(),
			securitypb.RateLimit_Action:      model.GetAction(),
			securitypb.RateLimit_TimeWindow:  model.GetTimeWindow(),
			securitypb.RateLimit_MaxRequests: model.GetMaxRequests(),
			securitypb.RateLimit_Timeout:     model.GetTimeout(),
		})

	_, err := r.storage.Exec(ctx, query)
	return err
}

func (r *RateLimit) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		securitypb.RateLimit_Id,
		securitypb.RateLimit_Name,
		securitypb.RateLimit_Description,
		securitypb.RateLimit_Expr,
		securitypb.RateLimit_Status,
		securitypb.RateLimit_Priority,
		securitypb.RateLimit_Action,
		securitypb.RateLimit_TimeWindow,
		securitypb.RateLimit_MaxRequests,
		securitypb.RateLimit_Timeout,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*securitypb.RateLimit, error) {
	var models []*securitypb.RateLimit
	for rows.Next() {
		model, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func scanOne(row dbx.Row) (*securitypb.RateLimit, error) {
	var (
		createdAt, updatedAt time.Time
		model                securitypb.RateLimit
	)

	err := row.Scan(
		&model.Id,
		&model.Name,
		&model.Description,
		&model.Expr,
		&model.Status,
		&model.Priority,
		&model.Action,
		&model.TimeWindow,
		&model.MaxRequests,
		&model.Timeout,
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
