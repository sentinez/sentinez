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

package secratelimit

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
type ISecRateLimit interface {
	Create(ctx context.Context, model *securitypb.SecRateLimit) (*securitypb.SecRateLimit, error)
	Update(ctx context.Context, model *securitypb.SecRateLimit, paths []string) error
	Get(ctx context.Context, id string) (*securitypb.SecRateLimit, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) ISecRateLimit

	List(ctx context.Context, req *securitypb.ListSecRateLimitsRequest) ([]*securitypb.SecRateLimit, int64, error)
}

var _ ISecRateLimit = (*SecRateLimit)(nil)

// New creates a SecRateLimit repository backed by Postgres.
func New(ctx context.Context,
	appConf *settingpb.Config) (ISecRateLimit, error) {
	storage, err := postgres.New[securitypb.SecRateLimit](ctx, appConf,
		schema.SecuritySecRateLimits())
	if err != nil {
		return nil, err
	}

	return &SecRateLimit{
		storage: storage,
	}, nil
}

// SecRateLimit is the Postgres repository for rate limit rules.
type SecRateLimit struct {
	storage dbx.Database[securitypb.SecRateLimit]
}

func (r *SecRateLimit) WithTX(tx *postgres.TxSession) ISecRateLimit {
	return &SecRateLimit{
		storage: postgres.WithTx(tx, r.storage),
	}
}

func buildListQuery(builder sq.SelectBuilder,
	req *securitypb.ListSecRateLimitsRequest) sq.SelectBuilder {
	if len(req.GetIds()) > 0 {
		builder = builder.Where(sq.Eq{securitypb.SecRateLimit_Id: req.GetIds()})
	}
	if ids := req.GetResourceIds(); len(ids) > 0 {
		builder = builder.Where(
			sq.Eq{securitypb.SecRateLimit_OwnerResourceId: ids})
	}
	return builder
}

func (r *SecRateLimit) List(ctx context.Context,
	req *securitypb.ListSecRateLimitsRequest,
) ([]*securitypb.SecRateLimit, int64, error) {
	builder := r.selectQuery(req.GetPage())
	builder = buildListQuery(builder, req)

	var total int64
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	szlog.Debug("[security.secratelimits] query: ", query, " args: ", args)

	models, err := r.storage.
		CollectRows(ctx, builder, securitypb.ScanSecRateLimit)
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
func (r *SecRateLimit) total(
	ctx context.Context, req *securitypb.ListSecRateLimitsRequest,
) (int64, error) {
	builder := buildListQuery(
		postgres.CountBuilder(r.storage), req)

	var total int64
	if err := r.storage.Query(ctx, builder, &total); err != nil {
		return 0, err
	}

	return total, nil
}

func (r *SecRateLimit) Create(ctx context.Context,
	model *securitypb.SecRateLimit) (*securitypb.SecRateLimit, error) {
	model.Id = szrand.NewUUID(szprefix.SecRateLimit)
	query := postgres.InsertBuilder(r.storage, rateLimitRow(model))

	if _, err := r.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return model, nil
}

func (r *SecRateLimit) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *SecRateLimit) Get(
	ctx context.Context,
	id string,
) (*securitypb.SecRateLimit, error) {
	builder := r.selectQuery(nil).
		Where(sq.Eq{securitypb.SecRateLimit_Id: id})
	return r.storage.Select(ctx, builder, securitypb.ScanOneSecRateLimit)
}

// Update writes the columns of the fields of model paths names, as
// szprotobuf.ApplyMask returns them; a path of no column is skipped.
func (r *SecRateLimit) Update(ctx context.Context,
	model *securitypb.SecRateLimit, paths []string) error {

	query := postgres.UpdateBuilder(r.storage, model.GetId()).
		SetMap(rateLimitRow(model).Only(paths))

	_, err := r.storage.Exec(ctx, query)
	return err
}

// rateLimitRow the stored columns of model.
func rateLimitRow(model *securitypb.SecRateLimit) postgres.M {
	owner := model.GetOwner()

	return postgres.M{
		securitypb.SecRateLimit_Id:              model.GetId(),
		securitypb.SecRateLimit_Name:            model.GetName(),
		securitypb.SecRateLimit_Description:     model.GetDescription(),
		securitypb.SecRateLimit_Expr:            model.GetExpr(),
		securitypb.SecRateLimit_Status:          model.GetStatus(),
		securitypb.SecRateLimit_Priority:        model.GetPriority(),
		securitypb.SecRateLimit_Action:          model.GetAction(),
		securitypb.SecRateLimit_TimeWindow:      model.GetTimeWindow(),
		securitypb.SecRateLimit_MaxRequests:     model.GetMaxRequests(),
		securitypb.SecRateLimit_Timeout:         model.GetTimeout(),
		securitypb.SecRateLimit_OwnerResourceId: owner.GetResourceId(),
		securitypb.SecRateLimit_OwnerTenantId:   owner.GetTenantId(),
	}
}

func (r *SecRateLimit) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		securitypb.SecRateLimitColumns()...)
}
