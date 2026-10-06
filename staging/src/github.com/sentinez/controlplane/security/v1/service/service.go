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

	"github.com/sentinez/controlplane/security/v1/repos/secrule"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/errorx"
)

var _ securitypb.SecurityServiceServer = (*SecurityService)(nil)

// New creates a new SecurityService.
func New(
	config *settingpb.Config,
	secRuleRepo secrule.ISecRule,
) *SecurityService {
	return &SecurityService{
		config:      config,
		secRuleRepo: secRuleRepo,
	}
}

// SecurityService handles security operations.
type SecurityService struct {
	config      *settingpb.Config
	secRuleRepo secrule.ISecRule
}

// mapSecRule converts a DB SecRule into an API SecRule response.
func (srv *SecurityService) mapSecRule(_ context.Context,
	dbRB *securitypb.SecRule,
) *securitypb.SecRule {
	if dbRB == nil {
		return nil
	}

	return &securitypb.SecRule{
		Id:          dbRB.GetId(),
		Name:        dbRB.GetName(),
		Description: dbRB.GetDescription(),
		Expr:        dbRB.GetExpr(),
		Action:      dbRB.GetAction(),
		Status:      dbRB.GetStatus(),
		Priority:    dbRB.GetPriority(),
		Metadata:    dbRB.GetMetadata(),
	}
}

// ── SecRule ────────────────────────────────────────────────────────────

// CreateSecRule creates a new WAF rule based
func (srv *SecurityService) CreateSecRule(ctx context.Context,
	req *securitypb.CreateSecRuleRequest,
) (*securitypb.CreateSecRuleResponse, error) {
	if req.GetSecRule() == nil {
		return nil, errorx.StatusInvalidArgumentF("sec_rule is required")
	}

	created, err := srv.secRuleRepo.Create(ctx, req.GetSecRule())
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateSecRuleResponse{Id: created.GetId()}, nil
}

func (srv *SecurityService) GetSecRule(ctx context.Context,
	req *securitypb.GetSecRuleRequest,
) (*securitypb.GetSecRuleResponse, error) {
	dbRB, err := srv.secRuleRepo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &securitypb.
		GetSecRuleResponse{SecRule: srv.mapSecRule(ctx, dbRB)}, nil
}

// UpdateSecRule updates an existing WAF rule based
func (srv *SecurityService) UpdateSecRule(ctx context.Context,
	req *securitypb.UpdateSecRuleRequest,
) (*securitypb.UpdateSecRuleResponse, error) {
	if req.GetSecRule() == nil {
		return nil, errorx.StatusInvalidArgumentF("sec_rule is required")
	}

	dbRB, err := srv.secRuleRepo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	// Validate rule structure

	dbRB.Name = req.GetSecRule().GetName()
	dbRB.Description = req.GetSecRule().GetDescription()
	dbRB.Expr = req.GetSecRule().GetExpr()
	dbRB.Status = req.GetSecRule().GetStatus()
	dbRB.Priority = req.GetSecRule().GetPriority()
	dbRB.Action = req.GetSecRule().GetAction()

	if err := srv.secRuleRepo.Update(ctx, dbRB); err != nil {
		return nil, err
	}

	return &securitypb.
		UpdateSecRuleResponse{SecRule: srv.mapSecRule(ctx, dbRB)}, nil
}

func (srv *SecurityService) DeleteSecRule(ctx context.Context,
	req *securitypb.DeleteSecRuleRequest,
) (*securitypb.DeleteSecRuleResponse, error) {
	if err := srv.secRuleRepo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &securitypb.DeleteSecRuleResponse{}, nil
}

func (srv *SecurityService) ListSecRules(ctx context.Context,
	req *securitypb.ListSecRulesRequest,
) (*securitypb.ListSecRulesResponse, error) {
	dbRBs, total, err := srv.secRuleRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	var rbs []*securitypb.SecRule
	for _, dbRB := range dbRBs {
		rbs = append(rbs, srv.mapSecRule(ctx, dbRB))
	}

	return &securitypb.ListSecRulesResponse{
		SecRules: rbs,
		Total:    total,
	}, nil
}

func (srv *SecurityService) Status(_ context.Context,
	_ *securitypb.StatusRequest,
) (*securitypb.StatusResponse, error) {
	return &securitypb.StatusResponse{Msg: "OK"}, nil
}
