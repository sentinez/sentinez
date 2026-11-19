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

package secruleset

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
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// The setting column holds the proto JSON of the CRS setting: fields by
// their proto name, enums by their value name. The list filters rely on
// this encoding.
var _marshalRuleset = protojson.MarshalOptions{UseProtoNames: true}

// nolint
type ISecRuleset interface {
	Create(ctx context.Context, model *securitypb.SecRuleset) (*securitypb.SecRuleset, error)
	Update(ctx context.Context, model *securitypb.SecRuleset, paths []string) error
	Get(ctx context.Context, id string) (*securitypb.SecRuleset, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) ISecRuleset

	List(ctx context.Context, req *securitypb.ListSecRulesetsRequest) ([]*securitypb.SecRuleset, int64, error)
}

var _ ISecRuleset = (*SecRuleset)(nil)

// New creates a SecRuleset repository backed by Postgres.
func New(
	ctx context.Context, appConf *settingpb.Config,
) (ISecRuleset, error) {
	storage, err := postgres.New[securitypb.SecRuleset](ctx, appConf,
		schema.SecuritySecRulesets())
	if err != nil {
		return nil, err
	}

	return &SecRuleset{
		storage: storage,
	}, nil
}

// SecRuleset is the Postgres repository for OWASP CRS configurations.
type SecRuleset struct {
	storage dbx.Database[securitypb.SecRuleset]
}

func (r *SecRuleset) WithTX(tx *postgres.TxSession) ISecRuleset {
	return &SecRuleset{
		storage: postgres.WithTx(tx, r.storage),
	}
}

func (r *SecRuleset) List(ctx context.Context,
	req *securitypb.ListSecRulesetsRequest,
) ([]*securitypb.SecRuleset, int64, error) {
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
	szlog.Debug("[security.secrulesets] query: ", query, " args: ", args)

	models, err := r.storage.CollectRows(
		ctx, builder, securitypb.ScanSecRuleset)
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
func (r *SecRuleset) total(
	ctx context.Context, conds []sq.Sqlizer,
) (int64, error) {
	builder := postgres.CountBuilder(r.storage)
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	var total int64
	if err := r.storage.Query(ctx, builder, &total); err != nil {
		return 0, err
	}

	return total, nil
}

func (r *SecRuleset) Create(ctx context.Context,
	model *securitypb.SecRuleset) (*securitypb.SecRuleset, error) {
	model.Id = szrand.NewUUID(szprefix.SecRuleset)
	row, err := rulesetRow(model)
	if err != nil {
		return nil, err
	}

	query := postgres.InsertBuilder(r.storage, row)

	if _, err := r.storage.Insert(ctx, query); err != nil {
		return nil, err
	}

	return model, nil
}

func (r *SecRuleset) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *SecRuleset) Get(
	ctx context.Context,
	id string,
) (*securitypb.SecRuleset, error) {
	builder := r.selectQuery(nil).
		Where(sq.Eq{securitypb.SecRuleset_Id: id})
	return r.storage.Select(ctx, builder, securitypb.ScanOneSecRuleset)
}

// Update writes the columns of the fields of model paths names, as
// szprotobuf.ApplyMask returns them; a path of no column is skipped.
func (r *SecRuleset) Update(ctx context.Context,
	model *securitypb.SecRuleset, paths []string) error {

	row, err := rulesetRow(model)
	if err != nil {
		return err
	}

	query := postgres.UpdateBuilder(r.storage, model.GetId()).
		SetMap(row.Only(paths))

	_, err = r.storage.Exec(ctx, query)
	return err
}

// rulesetRow the stored columns of model.
func rulesetRow(model *securitypb.SecRuleset) (postgres.M, error) {
	crsJSON, err := marshalRuleset(model.GetSetting())
	if err != nil {
		return nil, err
	}

	owner := model.GetOwner()

	return postgres.M{
		securitypb.SecRuleset_Id:              model.GetId(),
		securitypb.SecRuleset_Name:            model.GetName(),
		securitypb.SecRuleset_Description:     model.GetDescription(),
		securitypb.SecRuleset_Setting:         crsJSON,
		securitypb.SecRuleset_OwnerResourceId: owner.GetResourceId(),
		securitypb.SecRuleset_OwnerTenantId:   owner.GetTenantId(),
	}, nil
}

func (r *SecRuleset) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		securitypb.SecRulesetColumns()...)
}

// marshalRuleset a nil setting is stored as an empty object,
// so the setting filters never meet a NULL. An unset provider is stored
// as PROVIDER_OWASP, the one the edge runs it with, so the provider
// filter is a plain containment.
func marshalRuleset(setting *rulesetpb.Setting) (string, error) {
	if setting == nil {
		return "{}", nil
	}

	if setting.GetProvider() == rulesetpb.Provider_PROVIDER_UNSPECIFIED {
		setting = proto.CloneOf(setting)
		setting.Provider = rulesetpb.Provider_PROVIDER_OWASP
	}

	data, err := _marshalRuleset.Marshal(setting)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
