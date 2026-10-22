// Copyright 2026 Duc-Hung Ho.
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

package tenantrepo

import (
	"context"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/core/szcmn/sztables"
	"github.com/sentinez/core/szstorage/szdbx"
	"github.com/sentinez/core/szstorage/szdbx/szpostgres"
	"github.com/sentinez/core/szstorage/szstorageutils/sztable"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ ITenant = (*Tenants)(nil)

// _likeEscaper escapes the wildcards of a LIKE pattern.
var _likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// nolint
type ITenant interface {
	Create(ctx context.Context, tenant *tenantpb.Tenant) (*tenantpb.Tenant, error)
	Update(ctx context.Context, tenant *tenantpb.Tenant) error
	Get(ctx context.Context, id string) (*tenantpb.Tenant, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *szpostgres.TxSession) ITenant

	List(ctx context.Context, req *tenantpb.ListTenantRequest) (*tenantpb.ListTenantResponse, error)
}

func New(ctx context.Context, appConf *settingpb.Config) (ITenant, error) {
	storage, err := szpostgres.New[tenantpb.Tenant](ctx, appConf,
		szdbx.WithTable(sztables.Tenants),
		szdbx.WithColumns(szdbx.ColumnM{
			tenantpb.Tenant_Id:          szpostgres.String,
			tenantpb.Tenant_Name:        szpostgres.String,
			tenantpb.Tenant_Description: szpostgres.String,
			tenantpb.Tenant_Plan:        szpostgres.Int4,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Tenants{
		storage: storage,
	}, nil
}

type Tenants struct {
	storage szdbx.Database[tenantpb.Tenant]
}

func (t *Tenants) WithTX(tx *szpostgres.TxSession) ITenant {
	return &Tenants{
		storage: szpostgres.WithTx(tx, t.storage),
	}
}

// List implements ITenant.
func (t *Tenants) List(ctx context.Context,
	req *tenantpb.ListTenantRequest) (*tenantpb.ListTenantResponse, error) {

	conds := listConditions(req)

	builder := t.selectQuery(req.GetPage())
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	zlog.Debug("[tenants] query: ", query, " args: ", args)

	tenants, err := t.storage.CollectRows(ctx, builder, scan)
	if err != nil {
		return nil, err
	}

	var total int64
	if req.GetPage().GetTotal() {
		total, err = t.total(ctx, conds)
		if err != nil {
			return nil, err
		}
	}

	return &tenantpb.ListTenantResponse{Tenants: tenants, Total: total}, nil
}

// total counts the rows matching the list conditions.
func (t *Tenants) total(
	ctx context.Context, conds []sq.Sqlizer,
) (int64, error) {
	builder := sq.Select("COUNT(*)").From(t.storage.Table())
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	var total int64
	if err := t.storage.Query(ctx, builder, &total); err != nil {
		return 0, err
	}

	return total, nil
}

// listConditions maps the filters of the request to WHERE conditions,
// all of them must match.
func listConditions(req *tenantpb.ListTenantRequest) []sq.Sqlizer {
	var conds []sq.Sqlizer

	if req.GetPlan() != typepb.Plan_PLAN_UNSPECIFIED {
		conds = append(conds, sq.Eq{tenantpb.Tenant_Plan: req.GetPlan()})
	}

	if name := req.GetName(); name != "" {
		conds = append(conds, sq.ILike{
			tenantpb.Tenant_Name: "%" + _likeEscaper.Replace(name) + "%",
		})
	}

	return conds
}

// Create implements ITenant.
func (t *Tenants) Create(ctx context.Context,
	tenant *tenantpb.Tenant) (*tenantpb.Tenant, error) {

	tenant.Id = rand.NewId(sztable.NewPrimaryKey(sztables.Tenants))
	query := szpostgres.InsertBuilder(t.storage, szpostgres.M{
		tenantpb.Tenant_Id:          tenant.GetId(),
		tenantpb.Tenant_Name:        tenant.GetName(),
		tenantpb.Tenant_Description: tenant.GetDescription(),
		tenantpb.Tenant_Plan:        tenant.GetPlan(),
	})

	if _, err := t.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return tenant, nil
}

// Delete implements ITenant.
func (t *Tenants) Delete(ctx context.Context, id string) error {
	return t.storage.Delete(ctx, id)
}

// Get implements ITenant.
func (t *Tenants) Get(
	ctx context.Context, id string) (*tenantpb.Tenant, error) {

	builder := t.selectQuery(nil).Where(sq.Eq{tenantpb.Tenant_Id: id})
	return t.storage.Select(ctx, builder, scanOne)
}

// Update implements ITenant. Zero-valued fields are left unchanged.
func (t *Tenants) Update(ctx context.Context, tenant *tenantpb.Tenant) error {
	query := szpostgres.UpdateBuilder(t.storage, tenant.GetId())

	if tenant.GetName() != "" {
		query = query.Set(tenantpb.Tenant_Name, tenant.GetName())
	}
	if tenant.GetDescription() != "" {
		query = query.Set(
			tenantpb.Tenant_Description, tenant.GetDescription())
	}
	if tenant.GetPlan() != typepb.Plan_PLAN_UNSPECIFIED {
		query = query.Set(tenantpb.Tenant_Plan, tenant.GetPlan())
	}

	_, err := t.storage.Exec(ctx, query)
	return err
}

func (t *Tenants) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return szpostgres.SelectBuilder(t.storage, page,
		tenantpb.Tenant_Id,
		// rows created before these columns existed hold NULL
		szpostgres.SelectCoalesce(tenantpb.Tenant_Name),
		szpostgres.SelectCoalesce(tenantpb.Tenant_Description),
		tenantpb.Tenant_Plan,
		typepb.Metadata_CreatedAt,
		typepb.Metadata_UpdatedAt,
	)
}

func scan(rows szdbx.Rows) ([]*tenantpb.Tenant, error) {
	var tenants []*tenantpb.Tenant

	for rows.Next() {
		tenant, err := scanOne(rows)
		if err != nil {
			return nil, err
		}

		tenants = append(tenants, tenant)
	}

	return tenants, rows.Err()
}

func scanOne(row szdbx.Row) (*tenantpb.Tenant, error) {
	var (
		createdAt, updatedAt time.Time
		tenant               tenantpb.Tenant
	)

	err := row.Scan(
		&tenant.Id,
		&tenant.Name,
		&tenant.Description,
		&tenant.Plan,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	tenant.Metadata = &typepb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &tenant, nil
}
