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

package securitysvc

import (
	"context"

	"github.com/sentinez/controlplane/pkg/auditchange"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szprotobuf"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	"google.golang.org/protobuf/proto"
)

// mapRuleset converts a DB SecRuleset into an API SecRuleset
// response.
func (srv *SecurityService) mapRuleset(_ context.Context,
	dbRuleset *securitypb.SecRuleset,
) *securitypb.SecRuleset {
	if dbRuleset == nil {
		return nil
	}

	return &securitypb.SecRuleset{
		Id:          dbRuleset.GetId(),
		Name:        dbRuleset.GetName(),
		Description: dbRuleset.GetDescription(),
		Setting:     dbRuleset.GetSetting(),
		Owner:       dbRuleset.GetOwner(),
	}
}

// ── SecRuleset ────────────────────────────────────────────────────────────

// CreateSecRuleset creates a new OWASP CRS configuration
func (srv *SecurityService) CreateSecRuleset(ctx context.Context,
	req *securitypb.CreateSecRulesetRequest,
) (*securitypb.CreateSecRulesetResponse, error) {
	if err := validateRuleset(req.GetSecruleset()); err != nil {
		return nil, err
	}

	owner, err := srv.newOwner(ctx, req.GetSecruleset().GetOwner())
	if err != nil {
		return nil, err
	}

	req.GetSecruleset().Owner = owner

	created, err := srv.secRulesetRepo.Create(ctx, req.GetSecruleset())
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateSecRulesetResponse{Id: created.GetId()}, nil
}

func (srv *SecurityService) GetSecRuleset(ctx context.Context,
	req *securitypb.GetSecRulesetRequest,
) (*securitypb.GetSecRulesetResponse, error) {
	dbRuleset, err := owned(ctx, srv, srv.secRulesetRepo.Get, "ruleset",
		req.GetId(), authz.Read)
	if err != nil {
		return nil, err
	}

	return &securitypb.GetSecRulesetResponse{
		Secruleset: srv.mapRuleset(ctx, dbRuleset),
	}, nil
}

// _rulesetUpdatable the fields of a ruleset an update writes.
var _rulesetUpdatable = []string{
	securitypb.SecRuleset_Name, securitypb.SecRuleset_Description,
	securitypb.SecRuleset_Setting,
}

// UpdateSecRuleset updates the fields of an OWASP CRS configuration its
// update_mask names; the configuration they make must be valid.
func (srv *SecurityService) UpdateSecRuleset(ctx context.Context,
	req *securitypb.UpdateSecRulesetRequest,
) (*securitypb.UpdateSecRulesetResponse, error) {
	if req.GetSecruleset() == nil {
		return nil, szerror.StatusInvalidArgumentF(
			"secruleset is required")
	}

	dbRuleset, err := owned(ctx, srv, srv.secRulesetRepo.Get, "ruleset",
		req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	before := proto.CloneOf(dbRuleset)
	paths, err := szprotobuf.ApplyMask(dbRuleset, req.GetSecruleset(),
		req.GetUpdateMask(), _rulesetUpdatable...)
	if err != nil {
		return nil, err
	}

	if err = validateRuleset(dbRuleset); err != nil {
		return nil, err
	}

	if err = srv.secRulesetRepo.Update(ctx, dbRuleset, paths); err != nil {
		return nil, err
	}

	auditchange.Set(ctx, before, dbRuleset, paths)

	return &securitypb.UpdateSecRulesetResponse{
		Secruleset: srv.mapRuleset(ctx, dbRuleset),
	}, nil
}

func (srv *SecurityService) DeleteSecRuleset(ctx context.Context,
	req *securitypb.DeleteSecRulesetRequest,
) (*securitypb.DeleteSecRulesetResponse, error) {
	_, err := owned(ctx, srv, srv.secRulesetRepo.Get, "ruleset",
		req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	if err = srv.secRulesetRepo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &securitypb.DeleteSecRulesetResponse{}, nil
}

func (srv *SecurityService) ListSecRulesets(ctx context.Context,
	req *securitypb.ListSecRulesetsRequest,
) (*securitypb.ListSecRulesetsResponse, error) {
	req, ok, err := authz.Scoped(ctx, srv.authz, req)
	if err != nil || !ok {
		return &securitypb.ListSecRulesetsResponse{}, err
	}

	dbRulesets, total, err := srv.secRulesetRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	rulesets := make([]*securitypb.SecRuleset, 0, len(dbRulesets))
	for _, dbRuleset := range dbRulesets {
		rulesets = append(rulesets, srv.mapRuleset(ctx, dbRuleset))
	}

	return &securitypb.ListSecRulesetsResponse{
		Secrulesets: rulesets,
		Total:       total,
	}, nil
}

// validateRuleset checks the configuration carries a setting. The
// setting itself is validated by the edge when it builds the rulesets.
func validateRuleset(crs *securitypb.SecRuleset) error {
	if crs == nil {
		return szerror.StatusInvalidArgumentF("secruleset is required")
	}
	if crs.GetSetting() == nil {
		return szerror.StatusInvalidArgumentF("secruleset.setting is required")
	}
	return nil
}
