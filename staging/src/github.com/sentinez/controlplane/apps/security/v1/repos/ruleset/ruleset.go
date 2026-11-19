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

package ruleset

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
)

// The ruleset column holds the proto JSON of the CRS setting: fields by
// their proto name, enums by their value name. The list filters rely on
// this encoding.
var _marshalRuleset = protojson.MarshalOptions{UseProtoNames: true}

// nolint
type IRuleset interface {
	Create(ctx context.Context, model *securitypb.Ruleset) (*securitypb.Ruleset, error)
	Update(ctx context.Context, model *securitypb.Ruleset, paths []string) error
	Get(ctx context.Context, id string) (*securitypb.Ruleset, error)
	Delete(ctx context.Context, id string) error

	WithTX(tx *postgres.TxSession) IRuleset

	List(ctx context.Context, req *securitypb.ListRulesetsRequest) ([]*securitypb.Ruleset, int64, error)
}

var _ IRuleset = (*Ruleset)(nil)

// New creates a Ruleset repository backed by Postgres.
func New(
	ctx context.Context, appConf *settingpb.Config,
) (IRuleset, error) {
	storage, err := postgres.New[securitypb.Ruleset](ctx, appConf,
		schema.SecurityRulesets())
	if err != nil {
		return nil, err
	}

	return &Ruleset{
		storage: storage,
	}, nil
}

// Ruleset is the Postgres repository for OWASP CRS configurations.
type Ruleset struct {
	storage dbx.Database[securitypb.Ruleset]
}

func (r *Ruleset) WithTX(tx *postgres.TxSession) IRuleset {
	return &Ruleset{
		storage: postgres.WithTx(tx, r.storage),
	}
}

func (r *Ruleset) List(ctx context.Context,
	req *securitypb.ListRulesetsRequest,
) ([]*securitypb.Ruleset, int64, error) {
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
	szlog.Debug("[security.ruleset] query: ", query, " args: ", args)

	models, err := r.storage.CollectRows(
		ctx, builder, securitypb.ScanRuleset)
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
func (r *Ruleset) total(
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

func (r *Ruleset) Create(ctx context.Context,
	model *securitypb.Ruleset) (*securitypb.Ruleset, error) {
	model.Id = szrand.NewUUID(szprefix.Ruleset)
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

func (r *Ruleset) Delete(ctx context.Context, id string) error {
	return r.storage.Delete(ctx, id)
}

func (r *Ruleset) Get(
	ctx context.Context,
	id string,
) (*securitypb.Ruleset, error) {
	builder := r.selectQuery(nil).
		Where(sq.Eq{securitypb.Ruleset_Id: id})
	return r.storage.Select(ctx, builder, securitypb.ScanOneRuleset)
}

// Update writes the columns of the fields of model paths names, as
// szprotobuf.ApplyMask returns them; a path of no column is skipped.
func (r *Ruleset) Update(ctx context.Context,
	model *securitypb.Ruleset, paths []string) error {

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
func rulesetRow(model *securitypb.Ruleset) (postgres.M, error) {
	crsJSON, err := marshalRuleset(model.GetRuleset())
	if err != nil {
		return nil, err
	}

	owner := model.GetOwner()

	return postgres.M{
		securitypb.Ruleset_Id:              model.GetId(),
		securitypb.Ruleset_Name:            model.GetName(),
		securitypb.Ruleset_Description:     model.GetDescription(),
		securitypb.Ruleset_Ruleset:         crsJSON,
		securitypb.Ruleset_OwnerResourceId: owner.GetResourceId(),
		securitypb.Ruleset_OwnerTenantId:   owner.GetTenantId(),
	}, nil
}

func (r *Ruleset) selectQuery(page *typepb.Pages) sq.SelectBuilder {
	return postgres.SelectBuilder(r.storage, page,
		securitypb.RulesetColumns()...)
}

// marshalRuleset a nil setting is stored as an empty object,
// so the ruleset filters never meet a NULL.
func marshalRuleset(setting *rulesetpb.Ruleset) (string, error) {
	if setting == nil {
		return "{}", nil
	}

	data, err := _marshalRuleset.Marshal(setting)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
