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

	"github.com/sentinez/core/rules/builder"
	"github.com/sentinez/modules/security/v1/repos/rulebased"
	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/security/v1"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	"github.com/sentinez/shared/errorx"
)

var _ securitypb.SecurityServiceServer = (*SecurityService)(nil)

// New creates a new SecurityService.
func New(
	config *settingpb.Config,
	ruleBasedRepo rulebased.IRuleBased,
) *SecurityService {
	return &SecurityService{
		config:        config,
		ruleBasedRepo: ruleBasedRepo,
	}
}

// SecurityService handles security operations.
type SecurityService struct {
	config        *settingpb.Config
	ruleBasedRepo rulebased.IRuleBased
}

// toSystemNode converts a shared RuleGroup into a system rule node.
func (srv *SecurityService) toSystemNode(
	group *securitypb.RuleGroup) *rulepb.RuleBased_Node {
	if group == nil {
		return nil
	}

	var b *builder.GroupBuilder
	switch group.GetCombinator() {
	case rulepb.Logic_LOGIC_OR:
		b = builder.NewGroup(rulepb.Logic_LOGIC_OR)
	case rulepb.Logic_LOGIC_NOT:
		b = builder.NewGroup(rulepb.Logic_LOGIC_NOT)
	default:
		b = builder.NewGroup(rulepb.Logic_LOGIC_AND)
	}

	for _, node := range group.GetRules() {
		if node.GetGroup() != nil {
			subNode := srv.toSystemNode(node.GetGroup())
			if subNode != nil {
				b.AddGroup(&rulepb.RuleBased{Node: subNode})
			}
		} else if node.GetRule() != nil {
			r := node.GetRule()
			rb := builder.NewRule().WithCondition(
				r.GetField(), r.GetOperator(), r.GetValue(), r.GetKey())
			b.AddRule(rb.Build())
		}
	}

	res := b.Build()
	if group.GetNot() {
		return builder.Not(res).Build().GetNode()
	}

	return res.GetNode()
}

// mapRuleBased converts a DB RuleBased into an API RuleBased response.
func (srv *SecurityService) mapRuleBased(_ context.Context,
	dbRB *securitypb.RuleBased,
) *securitypb.RuleBased {
	if dbRB == nil {
		return nil
	}

	return &securitypb.RuleBased{
		Id:          dbRB.GetId(),
		Name:        dbRB.GetName(),
		Description: dbRB.GetDescription(),
		// This is now a RuleGroup pointer in the proto
		Node:     dbRB.GetNode(),
		Action:   dbRB.GetAction(),
		Status:   dbRB.GetStatus(),
		Priority: dbRB.GetPriority(),
		Metadata: dbRB.GetMetadata(),
	}
}

// ── RuleBased ────────────────────────────────────────────────────────────

// CreateRuleBased creates a new WAF rule based
func (srv *SecurityService) CreateRuleBased(ctx context.Context,
	req *securitypb.CreateRuleBasedRequest,
) (*securitypb.CreateRuleBasedResponse, error) {
	if req.GetRuleBased() == nil {
		return nil, errorx.StatusInvalidArgumentF("rule_based is required")
	}

	// Validate rule structure using the builder
	_ = srv.toSystemNode(req.GetRuleBased().GetNode())

	created, err := srv.ruleBasedRepo.Create(ctx, req.GetRuleBased())
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateRuleBasedResponse{Id: created.GetId()}, nil
}

func (srv *SecurityService) GetRuleBased(ctx context.Context,
	req *securitypb.GetRuleBasedRequest,
) (*securitypb.GetRuleBasedResponse, error) {
	dbRB, err := srv.ruleBasedRepo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &securitypb.
		GetRuleBasedResponse{RuleBased: srv.mapRuleBased(ctx, dbRB)}, nil
}

// UpdateRuleBased updates an existing WAF rule based
func (srv *SecurityService) UpdateRuleBased(ctx context.Context,
	req *securitypb.UpdateRuleBasedRequest,
) (*securitypb.UpdateRuleBasedResponse, error) {
	if req.GetRuleBased() == nil {
		return nil, errorx.StatusInvalidArgumentF("rule_based is required")
	}

	dbRB, err := srv.ruleBasedRepo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	// Validate rule structure
	_ = srv.toSystemNode(req.GetRuleBased().GetNode())

	dbRB.Name = req.GetRuleBased().GetName()
	dbRB.Description = req.GetRuleBased().GetDescription()
	dbRB.Node = req.GetRuleBased().GetNode()
	dbRB.Status = req.GetRuleBased().GetStatus()
	dbRB.Priority = req.GetRuleBased().GetPriority()
	dbRB.Action = req.GetRuleBased().GetAction()

	if err := srv.ruleBasedRepo.Update(ctx, dbRB); err != nil {
		return nil, err
	}

	return &securitypb.
		UpdateRuleBasedResponse{RuleBased: srv.mapRuleBased(ctx, dbRB)}, nil
}

func (srv *SecurityService) DeleteRuleBased(ctx context.Context,
	req *securitypb.DeleteRuleBasedRequest,
) (*securitypb.DeleteRuleBasedResponse, error) {
	if err := srv.ruleBasedRepo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &securitypb.DeleteRuleBasedResponse{}, nil
}

func (srv *SecurityService) ListRuleBaseds(ctx context.Context,
	req *securitypb.ListRuleBasedsRequest,
) (*securitypb.ListRuleBasedsResponse, error) {
	dbRBs, total, err := srv.ruleBasedRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	var rbs []*securitypb.RuleBased
	for _, dbRB := range dbRBs {
		rbs = append(rbs, srv.mapRuleBased(ctx, dbRB))
	}

	return &securitypb.ListRuleBasedsResponse{
		RuleBaseds: rbs,
		Total:      total,
	}, nil
}

func (srv *SecurityService) Status(_ context.Context,
	_ *securitypb.StatusRequest,
) (*securitypb.StatusResponse, error) {
	return &securitypb.StatusResponse{Msg: "OK"}, nil
}
