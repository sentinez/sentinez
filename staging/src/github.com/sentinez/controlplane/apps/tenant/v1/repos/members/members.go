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

package memberrepo

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szprefix"
	"github.com/sentinez/pkg/szrand"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

var _ IMember = (*Members)(nil)

// nolint
type IMember interface {
	Create(ctx context.Context, member *tenantpb.Member) (*tenantpb.Member, error)
	Update(ctx context.Context, member *tenantpb.Member) error
	Get(ctx context.Context, id string) (*tenantpb.Member, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IMember

	List(ctx context.Context, req *tenantpb.ListMemberRequest) (*tenantpb.ListMemberResponse, error)
}

func New(ctx context.Context, appConf *settingpb.Config) (IMember, error) {
	storage, err := postgres.New[tenantpb.Member](ctx, appConf,
		schema.TenantMembers())
	if err != nil {
		return nil, err
	}

	return &Members{
		storage: storage,
	}, nil
}

type Members struct {
	storage dbx.Database[tenantpb.Member]
}

func (m *Members) WithTX(tx *postgres.TxSession) IMember {
	return &Members{
		storage: postgres.WithTx(tx, m.storage),
	}
}

// listConditions maps the filters of the request to WHERE conditions,
// all of them must match.
func listConditions(req *tenantpb.ListMemberRequest) []sq.Sqlizer {
	var conds []sq.Sqlizer

	// a slice becomes tenant_id IN (...)
	if tenantIds := req.GetTenantIds(); len(tenantIds) != 0 {
		conds = append(conds, sq.Eq{tenantpb.Member_TenantId: tenantIds})
	}

	if userId := req.GetUserId(); userId != "" {
		conds = append(conds, sq.Eq{tenantpb.Member_UserId: userId})
	}

	return conds
}

// List implements IMember.
func (m *Members) List(ctx context.Context,
	req *tenantpb.ListMemberRequest) (*tenantpb.ListMemberResponse, error) {

	conds := listConditions(req)

	builder := m.selectQuery(req.GetPage())
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	szlog.Debug("[members] query: ", query, " args: ", args)

	members, err := m.storage.CollectRows(ctx, builder, tenantpb.ScanMember)
	if err != nil {
		return nil, err
	}

	var total int64
	if req.GetPage().GetTotal() {
		total, err = m.total(ctx, conds)
		if err != nil {
			return nil, err
		}
	}

	return &tenantpb.ListMemberResponse{Members: members, Total: total}, nil
}

// total counts the rows matching the list conditions.
func (m *Members) total(
	ctx context.Context, conds []sq.Sqlizer,
) (int64, error) {
	builder := sq.Select("COUNT(*)").From(m.storage.Table())
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	var total int64
	if err := m.storage.Query(ctx, builder, &total); err != nil {
		return 0, err
	}

	return total, nil
}

// Create implements IMember.
func (m *Members) Create(ctx context.Context,
	member *tenantpb.Member) (*tenantpb.Member, error) {

	member.Id = szrand.NewUUID(szprefix.Member)
	query := postgres.InsertBuilder(m.storage, postgres.M{
		tenantpb.Member_Id:       member.GetId(),
		tenantpb.Member_UserId:   member.GetUserId(),
		tenantpb.Member_TenantId: member.GetTenantId(),
		tenantpb.Member_Roles:    fromRoles(member.GetRoles()),
	})

	if _, err := m.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return member, nil
}

// Delete implements IMember.
func (m *Members) Delete(ctx context.Context, id string) error {
	return m.storage.Delete(ctx, id)
}

// Get implements IMember.
func (m *Members) Get(
	ctx context.Context, id string) (*tenantpb.Member, error) {

	builder := m.selectQuery(nil).Where(sq.Eq{tenantpb.Member_Id: id})
	return m.storage.Select(ctx, builder, tenantpb.ScanOneMember)
}

// Update implements IMember. Only the roles of a member can change.
func (m *Members) Update(ctx context.Context, member *tenantpb.Member) error {
	query := postgres.UpdateBuilder(m.storage, member.GetId())

	if len(member.GetRoles()) != 0 {
		query = query.Set(
			tenantpb.Member_Roles, fromRoles(member.GetRoles()))
	}

	_, err := m.storage.Exec(ctx, query)
	return err
}

func (m *Members) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(m.storage, page,
		tenantpb.MemberColumns()...)
}

// fromRoles converts roles to the INTEGER[] column value.
func fromRoles(roles []typepb.Role) []int32 {
	values := make([]int32, 0, len(roles))
	for _, role := range roles {
		values = append(values, int32(role))
	}

	return values
}
