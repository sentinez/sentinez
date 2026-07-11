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

package resourcerepo

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/core/common/tables"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/postgres"
	"github.com/sentinez/core/storage/utils/table"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/tenant/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/protobuf/protox"
	"github.com/sentinez/shared/rand"
)

func New(ctx context.Context, appConf *confpb.Config) (*Resources, error) {
	storage, err := postgres.New[tenantpb.Resource](ctx, appConf,
		dbx.WithTable(tables.TenantResources),
		dbx.WithColumns(dbx.ColumnM{
			tenantpb.Resource_Id:              postgres.String,
			tenantpb.Resource_ResourceSetting: postgres.ByteA,
			tenantpb.Resource_ResourceName:    postgres.String,
			tenantpb.Resource_ResourceDomain:  postgres.String,
			tenantpb.Resource_Plan:            postgres.Int4,
			tenantpb.Resource_Status:          postgres.Int4,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Resources{
		storage: storage,
	}, nil
}

var _ IResource = (*Resources)(nil)

// nolint
type IResource interface {
	List(ctx context.Context, req *tenantpb.ListResourceRequest) (*tenantpb.ListResourceResponse, error)
	Create(ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error)
	Update(ctx context.Context, rs *tenantpb.Resource) error
	Get(ctx context.Context, id string) (*tenantpb.Resource, error)
	Delete(ctx context.Context, id string) error
}

type Resources struct {
	storage dbx.Database[tenantpb.Resource]
}

// List implements IResource.
func (rsc *Resources) List(ctx context.Context,
	req *tenantpb.ListResourceRequest) (*tenantpb.ListResourceResponse, error) {

	q := rsc.selectQ(req.GetPage())
	if req.GetPlan() != typepb.Plan_PLAN_UNSPECIFIED {
		q = q.Where(sq.Eq{tenantpb.Resource_Plan: req.GetPlan()})
	}

	if req.GetStatus() != typepb.Status_STATUS_UNSPECIFIED {
		q = q.Where(sq.Eq{tenantpb.Resource_Status: req.GetStatus()})
	}

	if len(req.GetResourceDomain()) != 0 {
		q = q.Where(
			sq.Eq{tenantpb.Resource_ResourceDomain: req.GetResourceDomain()})
	}

	if len(req.GetResourceName()) != 0 {
		q = q.Where(
			sq.Eq{tenantpb.Resource_ResourceName: req.GetResourceName()})
	}

	resources, err := rsc.storage.CollectRows(ctx, q, scan)
	if err != nil {
		return nil, err
	}

	resp := &tenantpb.ListResourceResponse{Resources: resources}
	if req.GetPage().GetTotal() {
		resp.Total, err = rsc.storage.Total(ctx)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (rsc *Resources) Create(
	ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error) {

	rs.Id = rand.NewID(table.NewPrimaryKey(tables.TenantResources))

	st, _ := protox.Marshal(rs.GetResourceSetting())

	query := postgres.InsertBuilder(rsc.storage, postgres.M{
		tenantpb.Resource_Id:              rs.GetId(),
		tenantpb.Resource_Plan:            rs.GetPlan(),
		tenantpb.Resource_Status:          rs.GetStatus(),
		tenantpb.Resource_ResourceDomain:  rs.GetResourceDomain(),
		tenantpb.Resource_ResourceName:    rs.GetResourceName(),
		tenantpb.Resource_ResourceSetting: st,
	})

	_, err := rsc.storage.Insert(ctx, query)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func (rsc *Resources) Update(
	ctx context.Context, req *tenantpb.Resource) error {

	q := postgres.UpdateBuilder(rsc.storage, req.GetId())
	if req.GetPlan() != typepb.Plan_PLAN_UNSPECIFIED {
		q = q.Set(tenantpb.Resource_Plan, req.GetPlan())
	}

	if req.GetStatus() != typepb.Status_STATUS_UNSPECIFIED {
		q = q.Set(tenantpb.Resource_Status, req.GetStatus())
	}

	if len(req.GetResourceDomain()) != 0 {
		q = q.Set(tenantpb.Resource_ResourceDomain, req.GetResourceDomain())
	}

	if len(req.GetResourceName()) != 0 {
		q = q.Set(tenantpb.Resource_ResourceName, req.GetResourceName())
	}

	_, err := rsc.storage.Exec(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (rsc *Resources) Get(
	ctx context.Context, id string) (*tenantpb.Resource, error) {
	q := rsc.selectQ(nil).Where(sq.Eq{tenantpb.Resource_Id: id})

	return rsc.storage.CollectOneRow(ctx, q, scanOne)
}

func (rsc *Resources) Delete(ctx context.Context, id string) error {
	return rsc.storage.Delete(ctx, id)
}

func (rsc *Resources) selectQ(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(rsc.storage, page,
		tenantpb.Resource_Id,
		tenantpb.Resource_Plan,
		tenantpb.Resource_Status,
		tenantpb.Resource_ResourceName,
		tenantpb.Resource_ResourceDomain,
		tenantpb.Resource_ResourceSetting,
	)
}

func scan(rows dbx.Rows) ([]*tenantpb.Resource, error) {
	var resources []*tenantpb.Resource
	for rows.Next() {
		resource, err := scanOne(rows)
		if err != nil {
			return nil, err
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func scanOne(row dbx.Row) (*tenantpb.Resource, error) {
	var (
		rsc tenantpb.Resource
		st  []byte
	)

	err := row.Scan(
		&rsc.Id,
		&rsc.Plan,
		&rsc.Status,
		&rsc.ResourceName,
		&rsc.ResourceDomain,
		&st,
	)
	if err != nil {
		return nil, err
	}

	rsc.ResourceSetting = &edgepb.Setting{}
	if err = protox.Unmarshal(st, rsc.ResourceSetting); err != nil {
		return nil, err
	}

	return &rsc, nil
}
