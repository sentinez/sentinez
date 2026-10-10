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

	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	"github.com/sentinez/shared/errorx"
)

// mapCoreRuleset converts a DB CoreRuleset into an API CoreRuleset
// response.
func (srv *SecurityService) mapCoreRuleset(_ context.Context,
	dbCRS *securitypb.CoreRuleset,
) *securitypb.CoreRuleset {
	if dbCRS == nil {
		return nil
	}

	return &securitypb.CoreRuleset{
		Id:          dbCRS.GetId(),
		Name:        dbCRS.GetName(),
		Description: dbCRS.GetDescription(),
		Content:     dbCRS.GetContent(),
		Metadata:    dbCRS.GetMetadata(),
	}
}

// ── CoreRuleset ────────────────────────────────────────────────────────

// CreateCoreRuleset creates a new OWASP CRS configuration
func (srv *SecurityService) CreateCoreRuleset(ctx context.Context,
	req *securitypb.CreateCoreRulesetRequest,
) (*securitypb.CreateCoreRulesetResponse, error) {
	if err := validateCoreRuleset(req.GetCoreRuleset()); err != nil {
		return nil, err
	}

	created, err := srv.coreRulesetRepo.Create(ctx, req.GetCoreRuleset())
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateCoreRulesetResponse{Id: created.GetId()}, nil
}

func (srv *SecurityService) GetCoreRuleset(ctx context.Context,
	req *securitypb.GetCoreRulesetRequest,
) (*securitypb.GetCoreRulesetResponse, error) {
	dbCRS, err := srv.coreRulesetRepo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &securitypb.GetCoreRulesetResponse{
		CoreRuleset: srv.mapCoreRuleset(ctx, dbCRS),
	}, nil
}

// UpdateCoreRuleset updates an existing OWASP CRS configuration
func (srv *SecurityService) UpdateCoreRuleset(ctx context.Context,
	req *securitypb.UpdateCoreRulesetRequest,
) (*securitypb.UpdateCoreRulesetResponse, error) {
	crs := req.GetCoreRuleset()
	if err := validateCoreRuleset(crs); err != nil {
		return nil, err
	}

	dbCRS, err := srv.coreRulesetRepo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	dbCRS.Name = crs.GetName()
	dbCRS.Description = crs.GetDescription()
	dbCRS.Content = crs.GetContent()

	if err := srv.coreRulesetRepo.Update(ctx, dbCRS); err != nil {
		return nil, err
	}

	return &securitypb.UpdateCoreRulesetResponse{
		CoreRuleset: srv.mapCoreRuleset(ctx, dbCRS),
	}, nil
}

func (srv *SecurityService) DeleteCoreRuleset(ctx context.Context,
	req *securitypb.DeleteCoreRulesetRequest,
) (*securitypb.DeleteCoreRulesetResponse, error) {
	if err := srv.coreRulesetRepo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &securitypb.DeleteCoreRulesetResponse{}, nil
}

func (srv *SecurityService) ListCoreRulesets(ctx context.Context,
	req *securitypb.ListCoreRulesetsRequest,
) (*securitypb.ListCoreRulesetsResponse, error) {
	dbCRSs, total, err := srv.coreRulesetRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	crss := make([]*securitypb.CoreRuleset, 0, len(dbCRSs))
	for _, dbCRS := range dbCRSs {
		crss = append(crss, srv.mapCoreRuleset(ctx, dbCRS))
	}

	return &securitypb.ListCoreRulesetsResponse{
		CoreRulesets: crss,
		Total:        total,
	}, nil
}

// validateCoreRuleset checks the configuration carries a setting. The
// setting itself is validated by the edge when it builds the rulesets.
func validateCoreRuleset(crs *securitypb.CoreRuleset) error {
	if crs == nil {
		return errorx.StatusInvalidArgumentF("core_ruleset is required")
	}
	if crs.GetContent() == nil {
		return errorx.StatusInvalidArgumentF("content is required")
	}
	return nil
}
