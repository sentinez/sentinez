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
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/szprefix"
	"github.com/sentinez/shared/szrand"
	"github.com/sentinez/shared/szstorage/dbx"
	"github.com/sentinez/shared/szstorage/dbx/postgres"
	"github.com/sentinez/shared/sztable"
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

	WithTX(tx *postgres.TxSession) ITenant

	List(ctx context.Context, req *tenantpb.ListTenantRequest) (*tenantpb.ListTenantResponse, error)
}

func New(ctx context.Context, appConf *settingpb.Config) (ITenant, error) {
	storage, err := postgres.New[tenantpb.Tenant](ctx, appConf,
		dbx.WithTable(sztable.Tenants),
		dbx.WithColumns(dbx.ColumnM{
			tenantpb.Tenant_Id:          postgres.String,
			tenantpb.Tenant_Name:        postgres.String,
			tenantpb.Tenant_Description: postgres.String,
			tenantpb.Tenant_Plan:        postgres.Int4,
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
	storage dbx.Database[tenantpb.Tenant]
}

func (t *Tenants) WithTX(tx *postgres.TxSession) ITenant {
	return &Tenants{
		storage: postgres.WithTx(tx, t.storage),
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
	szlog.Debug("[tenants] query: ", query, " args: ", args)

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

	tenant.Id = szrand.NewUUID(szprefix.Tenant)
	query := postgres.InsertBuilder(t.storage, postgres.M{
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
	query := postgres.UpdateBuilder(t.storage, tenant.GetId())

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
	return postgres.SelectBuilder(t.storage, page,
		tenantpb.Tenant_Id,
		// rows created before these columns existed hold NULL
		postgres.SelectCoalesce(tenantpb.Tenant_Name),
		postgres.SelectCoalesce(tenantpb.Tenant_Description),
		tenantpb.Tenant_Plan,
		typepb.Metadata_CreatedAt,
		typepb.Metadata_UpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*tenantpb.Tenant, error) {
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

func scanOne(row dbx.Row) (*tenantpb.Tenant, error) {
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
