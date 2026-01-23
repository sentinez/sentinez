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
	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/internal/shared/tables"
	"github.com/sentinez/sentinez/pkg/common/protobuf/protox"
	"github.com/sentinez/sentinez/pkg/storage/dbx"
	"github.com/sentinez/sentinez/pkg/storage/dbx/postgres"
	"github.com/sentinez/sentinez/pkg/storage/utils/table"
	"github.com/sentinez/shared/ids"
)

func New(ctx context.Context, appConf *confpb.Config) (*Resources, error) {
	storage, err := postgres.New[tenantpb.Resource](ctx, appConf,
		dbx.WithTable(tables.Resources),
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
	Create(ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error)
	Update(ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error)
	Get(ctx context.Context, id string) (*tenantpb.Resource, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req *tenantpb.ListResourceRequest) (*tenantpb.ListResourceResponse, error)
}

type Resources struct {
	storage dbx.Database[tenantpb.Resource]
}

// List implements IResource.
func (rsc *Resources) List(ctx context.Context,
	req *tenantpb.ListResourceRequest) (*tenantpb.ListResourceResponse, error) {

	q := rsc.selectQ(req.GetPage())

	resources, err := rsc.storage.CollectRows(ctx, q, scan)
	if err != nil {
		return nil, err
	}

	resp := &tenantpb.ListResourceResponse{Resources: resources}
	if req.GetPage().GetTotal() {
		total, err := rsc.storage.Total(ctx)
		if err != nil {
			return nil, err
		}

		resp.Total = int32(total)
	}

	return resp, nil
}

func (rsc *Resources) Create(
	ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error) {

	rs.Id = ids.NewID(table.NewPrimaryKey(tables.Resources))

	st, _ := protox.Marshal(rs.GetResourceSetting())

	query := postgres.InsertBuilder(rsc.storage, postgres.M{
		tenantpb.Resource_Id:              rs.GetId(),
		tenantpb.Resource_Plan:            rs.GetPlan(),
		tenantpb.Resource_Status:          rs.GetStatus(),
		tenantpb.Resource_ResourceDomain:  rs.GetResourceDomain(),
		tenantpb.Resource_ResourceName:    rs.GetResourceName(),
		tenantpb.Resource_ResourceSetting: st,
	})

	rsc.storage.Insert(ctx, query)

	return rs, nil
}

func (rsc *Resources) Update(
	ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (rsc *Resources) Get(
	ctx context.Context, id string) (*tenantpb.Resource, error) {
	q := rsc.selectQ(nil).Where(sq.Eq{tenantpb.Resource_Id: id})

	return rsc.storage.CollectOneRow(ctx, q, scanOne)
}

func (rsc *Resources) Delete(ctx context.Context, id string) error {
	return rsc.storage.Delete(ctx, id)
}

func (rsc *Resources) selectQ(page *common.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(rsc.storage, page,
		string(tenantpb.Resource_Id),
		string(tenantpb.Resource_Plan),
		string(tenantpb.Resource_Status),
		string(tenantpb.Resource_ResourceName),
		string(tenantpb.Resource_ResourceDomain),
		string(tenantpb.Resource_ResourceSetting),
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
