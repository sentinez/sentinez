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
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szprefix"
	"github.com/sentinez/pkg/szprotobuf/protox"
	"github.com/sentinez/pkg/szrand"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

func New(ctx context.Context, appConf *settingpb.Config) (*Resources, error) {
	storage, err := postgres.New[tenantpb.Resource](ctx, appConf,
		schema.TenantResources())
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
	Upsert(ctx context.Context, rs *tenantpb.Resource) error
	Update(ctx context.Context, rs *tenantpb.Resource) error
	Get(ctx context.Context, id string) (*tenantpb.Resource, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IResource
}

type Resources struct {
	storage dbx.Database[tenantpb.Resource]
}

// List implements IResource.
func (rsc *Resources) List(ctx context.Context,
	req *tenantpb.ListResourceRequest) (*tenantpb.ListResourceResponse, error) {

	q := rsc.listQ(req)

	resources, err := rsc.storage.CollectRows(ctx, q, tenantpb.ScanResource)
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

// WithTX returns the repository running its queries in tx.
func (rsc *Resources) WithTX(tx *postgres.TxSession) IResource {
	return &Resources{storage: postgres.WithTx(tx, rsc.storage)}
}

// Upsert inserts rs under its own id, or replaces the stored row when the
// id exists.
func (rsc *Resources) Upsert(ctx context.Context, rs *tenantpb.Resource) error {
	row, err := resourceRow(rs)
	if err != nil {
		return err
	}

	_, err = rsc.storage.Insert(ctx, postgres.UpsertBuilder(rsc.storage, row))
	return err
}

func (rsc *Resources) Create(
	ctx context.Context, rs *tenantpb.Resource) (*tenantpb.Resource, error) {

	rs.Id = szrand.NewUUID(szprefix.TenantResource)

	row, err := resourceRow(rs)
	if err != nil {
		return nil, err
	}

	_, err = rsc.storage.Insert(ctx, postgres.InsertBuilder(rsc.storage, row))
	if err != nil {
		return nil, err
	}

	return rs, nil
}

// resourceRow is the stored columns of rs, whose server is stored under
// the id of rs.
func resourceRow(rs *tenantpb.Resource) (postgres.M, error) {
	if rs.GetServer() != nil {
		rs.Server.ResourceId = rs.GetId()
	}

	row, err := serverColumns(rs.GetServer())
	if err != nil {
		return nil, err
	}

	row[tenantpb.Resource_Id] = rs.GetId()
	row[tenantpb.Resource_TenantId] = rs.GetTenantId()
	row[tenantpb.Resource_Status] = rs.GetStatus()

	return row, nil
}

// serverColumns is the stored server column and the columns derived from
// it, written together so that they never disagree.
func serverColumns(srv *edgepb.Server) (postgres.M, error) {
	st, err := protox.Marshal(srv)
	if err != nil {
		return nil, err
	}

	return postgres.M{
		tenantpb.Resource_Server:    st,
		schema.ResourceServerName:   srv.GetName(),
		schema.ResourceServerDomain: srv.GetResourceDomain(),
	}, nil
}

// Update sets the status of rs unless unspecified, and replaces the
// stored server with the one of rs when set.
func (rsc *Resources) Update(
	ctx context.Context, rs *tenantpb.Resource) error {

	q := postgres.UpdateBuilder(rsc.storage, rs.GetId())
	if rs.GetStatus() != typepb.Status_STATUS_UNSPECIFIED {
		q = q.Set(tenantpb.Resource_Status, rs.GetStatus())
	}

	if rs.GetServer() != nil {
		rs.Server.ResourceId = rs.GetId()

		cols, err := serverColumns(rs.GetServer())
		if err != nil {
			return err
		}

		q = q.SetMap(cols)
	}

	_, err := rsc.storage.Exec(ctx, q)

	return err
}

func (rsc *Resources) Get(
	ctx context.Context, id string) (*tenantpb.Resource, error) {
	q := rsc.selectQ(nil).Where(sq.Eq{tenantpb.Resource_Id: id})

	return rsc.storage.CollectOneRow(ctx, q, tenantpb.ScanOneResource)
}

func (rsc *Resources) Delete(ctx context.Context, id string) error {
	return rsc.storage.Delete(ctx, id)
}

// listQ selects the page of resources matching the filters of req.
func (rsc *Resources) listQ(req *tenantpb.ListResourceRequest) sq.SelectBuilder {
	q := rsc.selectQ(req.GetPage())
	if req.GetStatus() != typepb.Status_STATUS_UNSPECIFIED {
		q = q.Where(sq.Eq{tenantpb.Resource_Status: req.GetStatus()})
	}

	if len(req.GetResourceDomain()) != 0 {
		q = q.Where(
			sq.Eq{schema.ResourceServerDomain: req.GetResourceDomain()})
	}

	if len(req.GetResourceName()) != 0 {
		q = q.Where(sq.Eq{schema.ResourceServerName: req.GetResourceName()})
	}

	if len(req.GetTenantIds()) != 0 {
		q = q.Where(sq.Eq{tenantpb.Resource_TenantId: req.GetTenantIds()})
	}

	return q
}

func (rsc *Resources) selectQ(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(rsc.storage, page,
		tenantpb.ResourceColumns()...)
}
