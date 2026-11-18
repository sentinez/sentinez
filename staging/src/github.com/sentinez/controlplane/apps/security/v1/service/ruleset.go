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

// mapRuleset converts a DB Ruleset into an API Ruleset
// response.
func (srv *SecurityService) mapRuleset(_ context.Context,
	dbRuleset *securitypb.Ruleset,
) *securitypb.Ruleset {
	if dbRuleset == nil {
		return nil
	}

	return &securitypb.Ruleset{
		Id:          dbRuleset.GetId(),
		Name:        dbRuleset.GetName(),
		Description: dbRuleset.GetDescription(),
		CoreRuleset: dbRuleset.GetCoreRuleset(),
		Owner:       dbRuleset.GetOwner(),
	}
}

// ── Ruleset ────────────────────────────────────────────────────────────

// CreateRuleset creates a new OWASP CRS configuration
func (srv *SecurityService) CreateRuleset(ctx context.Context,
	req *securitypb.CreateRulesetRequest,
) (*securitypb.CreateRulesetResponse, error) {
	if err := validateRuleset(req.GetRuleset()); err != nil {
		return nil, err
	}

	owner, err := srv.newOwner(ctx, req.GetRuleset().GetOwner())
	if err != nil {
		return nil, err
	}

	req.GetRuleset().Owner = owner

	created, err := srv.rulesetRepo.Create(ctx, req.GetRuleset())
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateRulesetResponse{Id: created.GetId()}, nil
}

func (srv *SecurityService) GetRuleset(ctx context.Context,
	req *securitypb.GetRulesetRequest,
) (*securitypb.GetRulesetResponse, error) {
	dbRuleset, err := owned(ctx, srv, srv.rulesetRepo.Get, "core ruleset",
		req.GetId(), authz.Read)
	if err != nil {
		return nil, err
	}

	return &securitypb.GetRulesetResponse{
		Ruleset: srv.mapRuleset(ctx, dbRuleset),
	}, nil
}

// _coreRulesetUpdatable the fields of a core ruleset an update writes.
var _coreRulesetUpdatable = []string{
	securitypb.Ruleset_Name, securitypb.Ruleset_Description,
	securitypb.Ruleset_CoreRuleset,
}

// UpdateRuleset updates the fields of an OWASP CRS configuration its
// update_mask names; the configuration they make must be valid.
func (srv *SecurityService) UpdateRuleset(ctx context.Context,
	req *securitypb.UpdateRulesetRequest,
) (*securitypb.UpdateRulesetResponse, error) {
	if req.GetRuleset() == nil {
		return nil, szerror.StatusInvalidArgumentF(
			"ruleset is required")
	}

	dbRuleset, err := owned(ctx, srv, srv.rulesetRepo.Get, "core ruleset",
		req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	before := proto.CloneOf(dbRuleset)
	paths, err := szprotobuf.ApplyMask(dbRuleset, req.GetRuleset(),
		req.GetUpdateMask(), _coreRulesetUpdatable...)
	if err != nil {
		return nil, err
	}

	if err = validateRuleset(dbRuleset); err != nil {
		return nil, err
	}

	if err = srv.rulesetRepo.Update(ctx, dbRuleset, paths); err != nil {
		return nil, err
	}

	auditchange.Set(ctx, before, dbRuleset, paths)

	return &securitypb.UpdateRulesetResponse{
		Ruleset: srv.mapRuleset(ctx, dbRuleset),
	}, nil
}

func (srv *SecurityService) DeleteRuleset(ctx context.Context,
	req *securitypb.DeleteRulesetRequest,
) (*securitypb.DeleteRulesetResponse, error) {
	_, err := owned(ctx, srv, srv.rulesetRepo.Get, "core ruleset",
		req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	if err = srv.rulesetRepo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &securitypb.DeleteRulesetResponse{}, nil
}

func (srv *SecurityService) ListRulesets(ctx context.Context,
	req *securitypb.ListRulesetsRequest,
) (*securitypb.ListRulesetsResponse, error) {
	req, ok, err := authz.Scoped(ctx, srv.authz, req)
	if err != nil || !ok {
		return &securitypb.ListRulesetsResponse{}, err
	}

	dbRulesets, total, err := srv.rulesetRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	rulesets := make([]*securitypb.Ruleset, 0, len(dbRulesets))
	for _, dbRuleset := range dbRulesets {
		rulesets = append(rulesets, srv.mapRuleset(ctx, dbRuleset))
	}

	return &securitypb.ListRulesetsResponse{
		Rulesets: rulesets,
		Total:    total,
	}, nil
}

// validateRuleset checks the configuration carries a setting. The
// setting itself is validated by the edge when it builds the rulesets.
func validateRuleset(crs *securitypb.Ruleset) error {
	if crs == nil {
		return szerror.StatusInvalidArgumentF("ruleset is required")
	}
	if crs.GetCoreRuleset() == nil {
		return szerror.StatusInvalidArgumentF("core_ruleset is required")
	}
	return nil
}
