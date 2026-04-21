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

	"google.golang.org/protobuf/encoding/protojson"

	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/security/v1"
	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	"github.com/sentinez/sentinez/internal/core/security/v1/repos/rulegroups"
	"github.com/sentinez/sentinez/internal/core/security/v1/repos/rules"
	"github.com/sentinez/sentinez/pkg/common/errorx"
)

var _ securitypb.SecurityServiceServer = (*SecurityService)(nil)

// New creates a new SecurityService.
func New(
	config *confpb.Config,
	rulesRepo rules.IRule,
	rulegroupsRepo rulegroups.IRuleGroup,
) *SecurityService {
	return &SecurityService{
		config:         config,
		rulesRepo:      rulesRepo,
		rulegroupsRepo: rulegroupsRepo,
	}
}

// SecurityService handles security operations.
type SecurityService struct {
	config         *confpb.Config
	rulesRepo      rules.IRule
	rulegroupsRepo rulegroups.IRuleGroup
}

// ── Helpers ───────────────────────────────────────────────────────────────

func mapRuleToDB(apiRule *ruleenginepb.Rule) (*securitypb.Rule, error) {
	b, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(apiRule)
	if err != nil {
		return nil, err
	}
	return &securitypb.Rule{
		Name:        apiRule.GetName(),
		Description: apiRule.GetDescription(),
		Rule:        string(b),
	}, nil
}

func mapDBToRule(dbRule *securitypb.Rule) (*ruleenginepb.Rule, error) {
	var rule ruleenginepb.Rule
	if dbRule.Rule != "" {
		if err := protojson.Unmarshal([]byte(dbRule.Rule), &rule); err != nil {
			return nil, err
		}
	}
	rule.Id = dbRule.Id
	rule.Name = dbRule.Name
	rule.Description = dbRule.Description
	if dbRule.Metadata != nil {
		rule.CreatedAt = dbRule.Metadata.CreatedAt
		rule.UpdatedAt = dbRule.Metadata.UpdatedAt
	}
	return &rule, nil
}

// mapDBToRuleGroup translates the DB RuleGroup into ruleengine RuleGroup
func (srv *SecurityService) mapDBToRuleGroup(_ context.Context,
	dbRG *securitypb.RuleGroup,
) (*ruleenginepb.RuleGroup, error) {
	if dbRG == nil {
		return nil, nil
	}

	var node ruleenginepb.RuleGroup_Node
	if dbRG.Node != "" {
		if err := protojson.Unmarshal([]byte(dbRG.Node), &node); err != nil {
			return nil, err
		}
	}

	return &ruleenginepb.RuleGroup{
		Node: &node,
	}, nil
}

// ── RuleGroup ────────────────────────────────────────────────────────────

func (srv *SecurityService) CreateRuleGroup(ctx context.Context,
	req *securitypb.CreateRuleGroupRequest,
) (*securitypb.CreateRuleGroupResponse, error) {
	if req.RuleGroup == nil {
		return nil, errorx.StatusInvalidArgumentF("rule_group is required")
	}

	var nodeStr string
	if req.RuleGroup.Node != nil {
		b, err := protojson.MarshalOptions{EmitUnpopulated: true}.
			Marshal(req.RuleGroup.Node)
		if err != nil {
			return nil, err
		}
		nodeStr = string(b)
	}

	dbRG := &securitypb.RuleGroup{
		Name:        req.RuleGroup.Name,
		Description: req.RuleGroup.Description,
		Node:        nodeStr,
	}

	created, err := srv.rulegroupsRepo.Create(ctx, dbRG)
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateRuleGroupResponse{Id: created.Id}, nil
}

func (srv *SecurityService) GetRuleGroup(ctx context.Context,
	req *securitypb.GetRuleGroupRequest,
) (*securitypb.GetRuleGroupResponse, error) {
	dbRG, err := srv.rulegroupsRepo.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	rg, err := srv.mapDBToRuleGroup(ctx, dbRG)
	if err != nil {
		return nil, err
	}

	// Transfer top-level metadata from DB model to ruleenginepb structure
	if rg != nil {
		rg.Id = dbRG.Id
		rg.Name = dbRG.Name
		rg.Description = dbRG.Description
	}

	return &securitypb.GetRuleGroupResponse{RuleGroup: rg}, nil
}

//nolint:funlen
func (srv *SecurityService) UpdateRuleGroup(ctx context.Context,
	req *securitypb.UpdateRuleGroupRequest,
) (*securitypb.UpdateRuleGroupResponse, error) {
	if req.RuleGroup == nil {
		return nil, errorx.StatusInvalidArgumentF("rule_group is required")
	}

	dbRG, err := srv.rulegroupsRepo.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	dbRG.Name = req.RuleGroup.Name
	dbRG.Description = req.RuleGroup.Description

	if req.RuleGroup.Node != nil {
		b, err := protojson.MarshalOptions{EmitUnpopulated: true}.
			Marshal(req.RuleGroup.Node)
		if err != nil {
			return nil, err
		}
		dbRG.Node = string(b)
	} else {
		dbRG.Node = ""
	}

	if err := srv.rulegroupsRepo.Update(ctx, dbRG); err != nil {
		return nil, err
	}

	rg, err := srv.mapDBToRuleGroup(ctx, dbRG)
	if err != nil {
		return nil, err
	}

	if rg != nil {
		rg.Id = dbRG.Id
		rg.Name = dbRG.Name
		rg.Description = dbRG.Description
	}

	return &securitypb.UpdateRuleGroupResponse{RuleGroup: rg}, nil
}

func (srv *SecurityService) DeleteRuleGroup(ctx context.Context,
	req *securitypb.DeleteRuleGroupRequest,
) (*securitypb.DeleteRuleGroupResponse, error) {
	if err := srv.rulegroupsRepo.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &securitypb.DeleteRuleGroupResponse{}, nil
}

func (srv *SecurityService) ListRuleGroups(ctx context.Context,
	req *securitypb.ListRuleGroupsRequest,
) (*securitypb.ListRuleGroupsResponse, error) {
	dbRGs, total, err := srv.rulegroupsRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	var rgs []*ruleenginepb.RuleGroup
	for _, dbRG := range dbRGs {
		rg, err := srv.mapDBToRuleGroup(ctx, dbRG)
		if err != nil {
			return nil, err
		}
		if rg != nil {
			rg.Id = dbRG.Id
			rg.Name = dbRG.Name
			rg.Description = dbRG.Description
			rgs = append(rgs, rg)
		}
	}

	return &securitypb.ListRuleGroupsResponse{
		RuleGroups: rgs,
		Total:      total,
	}, nil
}

// ── Rule ──────────────────────────────────────────────────────────────────

func (srv *SecurityService) CreateRule(ctx context.Context,
	req *securitypb.CreateRuleRequest,
) (*securitypb.CreateRuleResponse, error) {
	if req.Rule == nil {
		return nil, errorx.StatusInvalidArgumentF("rule is required")
	}

	dbRule, err := mapRuleToDB(req.Rule)
	if err != nil {
		return nil, err
	}

	created, err := srv.rulesRepo.Create(ctx, dbRule)
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateRuleResponse{Id: created.Id}, nil
}

func (srv *SecurityService) GetRule(ctx context.Context,
	req *securitypb.GetRuleRequest,
) (*securitypb.GetRuleResponse, error) {
	dbRule, err := srv.rulesRepo.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	rule, err := mapDBToRule(dbRule)
	if err != nil {
		return nil, err
	}

	return &securitypb.GetRuleResponse{Rule: rule}, nil
}

func (srv *SecurityService) UpdateRule(ctx context.Context,
	req *securitypb.UpdateRuleRequest,
) (*securitypb.UpdateRuleResponse, error) {
	if req.Rule == nil {
		return nil, errorx.StatusInvalidArgumentF("rule is required")
	}

	dbRule, err := srv.rulesRepo.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	newDbRule, err := mapRuleToDB(req.Rule)
	if err != nil {
		return nil, err
	}

	dbRule.Name = newDbRule.Name
	dbRule.Description = newDbRule.Description
	dbRule.Rule = newDbRule.Rule

	if err := srv.rulesRepo.Update(ctx, dbRule); err != nil {
		return nil, err
	}

	rule, err := mapDBToRule(dbRule)
	if err != nil {
		return nil, err
	}

	return &securitypb.UpdateRuleResponse{Rule: rule}, nil
}

func (srv *SecurityService) DeleteRule(ctx context.Context,
	req *securitypb.DeleteRuleRequest,
) (*securitypb.DeleteRuleResponse, error) {
	if err := srv.rulesRepo.Delete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &securitypb.DeleteRuleResponse{}, nil
}

func (srv *SecurityService) ListRules(ctx context.Context,
	req *securitypb.ListRulesRequest,
) (*securitypb.ListRulesResponse, error) {
	dbRules, total, err := srv.rulesRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	var rules []*ruleenginepb.Rule
	for _, dbRule := range dbRules {
		rule, err := mapDBToRule(dbRule)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return &securitypb.ListRulesResponse{Rules: rules, Total: total}, nil
}

// ── Status ────────────────────────────────────────────────────────────────

func (srv *SecurityService) Status(_ context.Context,
	_ *securitypb.StatusRequest,
) (*securitypb.StatusResponse, error) {
	return &securitypb.StatusResponse{Msg: "OK"}, nil
}
