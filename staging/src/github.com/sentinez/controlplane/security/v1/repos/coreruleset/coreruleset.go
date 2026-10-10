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

package coreruleset

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/core/common/tables"
	"github.com/sentinez/core/storage/dbx"
	"github.com/sentinez/core/storage/dbx/postgres"
	"github.com/sentinez/core/storage/utils/table"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// The content column holds the proto JSON of the CRS setting: fields by
// their proto name, enums by their value name. The list filters rely on
// this encoding.
var (
	_marshalContent   = protojson.MarshalOptions{UseProtoNames: true}
	_unmarshalContent = protojson.UnmarshalOptions{DiscardUnknown: true}
)

// nolint
type ICoreRuleset interface {
	Create(ctx context.Context, model *securitypb.CoreRuleset) (*securitypb.CoreRuleset, error)
	Update(ctx context.Context, model *securitypb.CoreRuleset) error
	Get(ctx context.Context, id string) (*securitypb.CoreRuleset, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) ICoreRuleset

	List(ctx context.Context, req *securitypb.ListCoreRulesetsRequest) ([]*securitypb.CoreRuleset, int64, error)
}

var _ ICoreRuleset = (*CoreRuleset)(nil)

// New creates a CoreRuleset repository backed by Postgres.
func New(
	ctx context.Context, appConf *settingpb.Config,
) (ICoreRuleset, error) {
	storage, err := postgres.New[securitypb.CoreRuleset](ctx, appConf,
		dbx.WithTable(tables.SecurityCoreRulesets),
		dbx.WithColumns(dbx.ColumnM{
			securitypb.CoreRuleset_Id:          postgres.String,
			securitypb.CoreRuleset_Name:        postgres.String,
			securitypb.CoreRuleset_Description: postgres.String,
			securitypb.CoreRuleset_Content:     postgres.JSONB,
		}),
		// The list filters on content are JSONB containments.
		dbx.WithIndexes(dbx.Index{
			Column: securitypb.CoreRuleset_Content,
			Method: postgres.GIN,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &CoreRuleset{
		storage: storage,
	}, nil
}

// CoreRuleset is the Postgres repository for OWASP CRS configurations.
type CoreRuleset struct {
	storage dbx.Database[securitypb.CoreRuleset]
}

func (r *CoreRuleset) WithTX(tx *postgres.TxSession) ICoreRuleset {
	return &CoreRuleset{
		storage: postgres.WithTx(tx, r.storage),
	}
}

func (r *CoreRuleset) List(ctx context.Context,
	req *securitypb.ListCoreRulesetsRequest,
) ([]*securitypb.CoreRuleset, int64, error) {
	conds, err := listConditions(req)
	if err != nil {
		return nil, 0, err
	}

	builder := r.selectQuery(req.GetPage())
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	zlog.Debug("[security.coreruleset] query: ", query, " args: ", args)

	models, err := r.storage.CollectRows(ctx, builder, scan)
	if err != nil {
		return nil, 0, err
	}

	if !req.GetPage().GetTotal() {
		return models, 0, nil
	}

	total, err := r.total(ctx, conds)
	if err != nil {
		return nil, 0, err
	}

	return models, total, nil
}

// total counts the rows matching the list conditions.
func (r *CoreRuleset) total(
	ctx context.Context, conds []sq.Sqlizer,
) (int64, error) {
	builder := sq.Select("COUNT(*)").From(r.storage.Table())
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	var total int64
	if err := r.storage.Query(ctx, builder, &total); err != nil {
		return 0, err
	}

	return total, nil
}

func (r *CoreRuleset) Create(ctx context.Context,
	model *securitypb.CoreRuleset) (*securitypb.CoreRuleset, error) {
	content, err := marshalContent(model.GetContent())
	if err != nil {
		return nil, err
	}

	model.Id = rand.NewID(table.NewPrimaryKey(tables.SecurityCoreRulesets))
	query := postgres.InsertBuilder(r.storage, postgres.M{
		securitypb.CoreRuleset_Id:          model.GetId(),
		securitypb.CoreRuleset_Name:        model.GetName(),
		securitypb.CoreRuleset_Description: model.GetDescription(),
		securitypb.CoreRuleset_Content:     content,
	})

	if _, err := r.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return model, nil
}

func (r *CoreRuleset) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *CoreRuleset) Get(
	ctx context.Context,
	id string,
) (*securitypb.CoreRuleset, error) {
	builder := r.selectQuery(nil).
		Where(sq.Eq{securitypb.CoreRuleset_Id: id})
	return r.storage.Select(ctx, builder, scanOne)
}

func (r *CoreRuleset) Update(
	ctx context.Context,
	model *securitypb.CoreRuleset,
) error {
	content, err := marshalContent(model.GetContent())
	if err != nil {
		return err
	}

	query := postgres.UpdateBuilder(r.storage, model.GetId()).
		SetMap(sq.Eq{
			securitypb.CoreRuleset_Name:        model.GetName(),
			securitypb.CoreRuleset_Description: model.GetDescription(),
			securitypb.CoreRuleset_Content:     content,
		})

	_, err = r.storage.Exec(ctx, query)
	return err
}

func (r *CoreRuleset) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		securitypb.CoreRuleset_Id,
		securitypb.CoreRuleset_Name,
		securitypb.CoreRuleset_Description,
		securitypb.CoreRuleset_Content,
		dbx.FieldCreatedAt,
		dbx.FieldUpdatedAt,
	)
}

// marshalContent a setting without content is stored as an empty object,
// so the content filters never meet a NULL.
func marshalContent(content *corerulesetpb.CoreRuleset) (string, error) {
	if content == nil {
		return "{}", nil
	}

	data, err := _marshalContent.Marshal(content)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func scan(rows dbx.Rows) ([]*securitypb.CoreRuleset, error) {
	var models []*securitypb.CoreRuleset
	for rows.Next() {
		model, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, rows.Err()
}

func scanOne(row dbx.Row) (*securitypb.CoreRuleset, error) {
	var (
		createdAt, updatedAt time.Time
		content              []byte
		model                securitypb.CoreRuleset
	)

	err := row.Scan(
		&model.Id,
		&model.Name,
		&model.Description,
		&content,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(content) > 0 {
		model.Content = &corerulesetpb.CoreRuleset{}

		err = _unmarshalContent.Unmarshal(content, model.Content)
		if err != nil {
			return nil, err
		}
	}

	model.Metadata = &typepb.Metadata{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}

	return &model, nil
}
