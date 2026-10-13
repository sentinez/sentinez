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
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/core/common/tables"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/postgres"
	"github.com/sentinez/core/storage/utils/table"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		dbx.WithTable(tables.TenantMembers),
		dbx.WithColumns(dbx.ColumnM{
			tenantpb.Member_Id:       postgres.String,
			tenantpb.Member_UserId:   postgres.String,
			tenantpb.Member_TenantId: postgres.String,
			tenantpb.Member_Roles:    postgres.IntArr,
		}),
	)
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

func buildListQuery(builder sq.SelectBuilder,
	req *tenantpb.ListMemberRequest) sq.SelectBuilder {

	if req.GetTenantId() != "" {
		builder = builder.Where(
			sq.Eq{tenantpb.Member_TenantId: req.GetTenantId()})
	}

	if req.GetUserId() != "" {
		builder = builder.Where(
			sq.Eq{tenantpb.Member_UserId: req.GetUserId()})
	}

	return builder
}

// List implements IMember.
func (m *Members) List(ctx context.Context,
	req *tenantpb.ListMemberRequest) (*tenantpb.ListMemberResponse, error) {

	builder := buildListQuery(m.selectQuery(req.GetPage()), req)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	zlog.Debug("[members] query: ", query, " args: ", args)

	members, err := m.storage.CollectRows(ctx, builder, scan)
	if err != nil {
		return nil, err
	}

	var total int64
	if req.GetPage().GetTotal() {
		total, err = m.storage.Total(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &tenantpb.ListMemberResponse{Members: members, Total: total}, nil
}

// Create implements IMember.
func (m *Members) Create(ctx context.Context,
	member *tenantpb.Member) (*tenantpb.Member, error) {

	member.Id = rand.NewId(table.NewPrimaryKey(tables.TenantMembers))
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
	return m.storage.Select(ctx, builder, scanOne)
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
		tenantpb.Member_Id,
		tenantpb.Member_UserId,
		tenantpb.Member_TenantId,
		tenantpb.Member_Roles,
		typepb.Metadata_CreatedAt,
		typepb.Metadata_UpdatedAt,
	)
}

func scan(rows dbx.Rows) ([]*tenantpb.Member, error) {
	var members []*tenantpb.Member

	for rows.Next() {
		member, err := scanOne(rows)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	return members, rows.Err()
}

func scanOne(row dbx.Row) (*tenantpb.Member, error) {
	var (
		createdAt, updatedAt time.Time
		member               tenantpb.Member
		roles                []int32
	)

	err := row.Scan(
		&member.Id,
		&member.UserId,
		&member.TenantId,
		&roles,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	member.Roles = toRoles(roles)
	member.Metadata = &typepb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &member, nil
}

// fromRoles converts roles to the INTEGER[] column value.
func fromRoles(roles []typepb.Role) []int32 {
	values := make([]int32, 0, len(roles))
	for _, role := range roles {
		values = append(values, int32(role))
	}

	return values
}

func toRoles(values []int32) []typepb.Role {
	roles := make([]typepb.Role, 0, len(values))
	for _, value := range values {
		roles = append(roles, typepb.Role(value))
	}

	return roles
}
