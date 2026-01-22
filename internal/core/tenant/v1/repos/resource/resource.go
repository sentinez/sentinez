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

	tenantpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/tenant/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/sentinez/internal/shared/tables"
	"github.com/sentinez/sentinez/pkg/common/protobuf/protox"
	"github.com/sentinez/sentinez/pkg/storage/dbx"
	"github.com/sentinez/sentinez/pkg/storage/dbx/postgres"
	"github.com/sentinez/sentinez/pkg/storage/utils/table"
	"github.com/sentinez/shared/ids"
)

func New(ctx context.Context, appConf *confpb.Config) (*Resource, error) {
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

	return &Resource{
		storage: storage,
	}, nil
}

var _ IResource = (*Resource)(nil)

// nolint
type IResource interface {
	Create(ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error)
	Update(ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error)
	Get(ctx context.Context, id string) (*tenantpb.Resource, error)
	Delete(ctx context.Context, id string) error
}

type Resource struct {
	storage dbx.Database[tenantpb.Resource]
}

func (rsc *Resource) Create(
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

func (rsc *Resource) Update(
	ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (rsc *Resource) Get(
	ctx context.Context, id string) (*tenantpb.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (rsc *Resource) Delete(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
