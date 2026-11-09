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

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szprefix"
	"github.com/sentinez/pkg/szrand"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
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
		schema.SecurityRateLimits())
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
	if ids := req.GetResourceIds(); len(ids) > 0 {
		builder = builder.Where(
			sq.Eq{securitypb.RateLimit_OwnerResourceId: ids})
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
	szlog.Debug("[security.ratelimits] query: ", query, " args: ", args)

	models, err := r.storage.CollectRows(ctx, builder, securitypb.ScanRateLimit)
	if err != nil {
		return nil, 0, err
	}

	if req.GetPage().GetTotal() {
		total, err = r.total(ctx, req)
		if err != nil {
			return nil, 0, err
		}
	}

	return models, total, nil
}

// total counts the rows matching the list filters.
func (r *RateLimit) total(
	ctx context.Context, req *securitypb.ListRateLimitsRequest,
) (int64, error) {
	builder := buildListQuery(
		sq.Select("COUNT(*)").From(r.storage.Table()), req)

	var total int64
	if err := r.storage.Query(ctx, builder, &total); err != nil {
		return 0, err
	}

	return total, nil
}

func (r *RateLimit) Create(ctx context.Context,
	model *securitypb.RateLimit) (*securitypb.RateLimit, error) {
	model.Id = szrand.NewUUID(szprefix.RateLimit)
	query := postgres.InsertBuilder(r.storage, postgres.M{
		securitypb.RateLimit_Id:              model.GetId(),
		securitypb.RateLimit_Name:            model.GetName(),
		securitypb.RateLimit_Description:     model.GetDescription(),
		securitypb.RateLimit_Expr:            model.GetExpr(),
		securitypb.RateLimit_Status:          model.GetStatus(),
		securitypb.RateLimit_Priority:        model.GetPriority(),
		securitypb.RateLimit_Action:          model.GetAction(),
		securitypb.RateLimit_TimeWindow:      model.GetTimeWindow(),
		securitypb.RateLimit_MaxRequests:     model.GetMaxRequests(),
		securitypb.RateLimit_Timeout:         model.GetTimeout(),
		securitypb.RateLimit_OwnerResourceId: model.GetOwner().GetResourceId(),
		securitypb.RateLimit_OwnerTenantId:   model.GetOwner().GetTenantId(),
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
	return r.storage.Select(ctx, builder, securitypb.ScanOneRateLimit)
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
		securitypb.RateLimitColumns()...)
}
