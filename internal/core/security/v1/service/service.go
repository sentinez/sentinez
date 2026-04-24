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
	"github.com/sentinez/sentinez/internal/core/security/v1/repos/rulebased"
	"github.com/sentinez/sentinez/pkg/common/errorx"
)

var _ securitypb.SecurityServiceServer = (*SecurityService)(nil)

// New creates a new SecurityService.
func New(
	config *confpb.Config,
	ruleBasedRepo rulebased.IRuleBased,
) *SecurityService {
	return &SecurityService{
		config:        config,
		ruleBasedRepo: ruleBasedRepo,
	}
}

// SecurityService handles security operations.
type SecurityService struct {
	config        *confpb.Config
	ruleBasedRepo rulebased.IRuleBased
}

// mapDBToRuleBased translates the DB RuleBased into ruleengine RuleBased
func (srv *SecurityService) mapDBToRuleBased(_ context.Context,
	dbRB *securitypb.RuleBased,
) (*ruleenginepb.RuleBased, error) {
	if dbRB == nil {
		return nil, nil
	}

	var node ruleenginepb.RuleBased_Node
	if dbRB.GetNode() != "" {
		if err := protojson.
			Unmarshal([]byte(dbRB.GetNode()), &node); err != nil {
			return nil, err
		}
	}

	var action ruleenginepb.Action
	if dbRB.GetAction() != "" {
		if err := protojson.
			Unmarshal([]byte(dbRB.GetAction()), &action); err != nil {
			return nil, err
		}
	}

	return &ruleenginepb.RuleBased{
		Node:        &node,
		Action:      &action,
		Id:          dbRB.GetId(),
		Name:        dbRB.GetName(),
		Description: dbRB.GetDescription(),
		Priority:    dbRB.GetPriority(),
		Status:      dbRB.GetStatus(),
	}, nil
}

// ── RuleBased ────────────────────────────────────────────────────────────

// nolint:funlen
func (srv *SecurityService) CreateRuleBased(ctx context.Context,
	req *securitypb.CreateRuleBasedRequest,
) (*securitypb.CreateRuleBasedResponse, error) {
	if req.GetRuleBased() == nil {
		return nil, errorx.StatusInvalidArgumentF("rule_based is required")
	}

	var nodeStr string
	if req.GetRuleBased().GetNode() != nil {
		b, err := protojson.MarshalOptions{EmitUnpopulated: true}.
			Marshal(req.GetRuleBased().GetNode())
		if err != nil {
			return nil, err
		}
		nodeStr = string(b)
	}

	var action string
	if req.GetRuleBased().GetAction() != nil {
		b, err := protojson.MarshalOptions{EmitUnpopulated: true}.
			Marshal(req.GetRuleBased().GetAction())
		if err != nil {
			return nil, err
		}
		action = string(b)
	}

	dbRB := &securitypb.RuleBased{
		Name:        req.GetRuleBased().GetName(),
		Description: req.GetRuleBased().GetDescription(),
		Node:        nodeStr,
		Action:      action,
		Priority:    req.GetRuleBased().GetPriority(),
		Status:      req.GetRuleBased().GetStatus(),
	}

	created, err := srv.ruleBasedRepo.Create(ctx, dbRB)
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

	rb, err := srv.mapDBToRuleBased(ctx, dbRB)
	if err != nil {
		return nil, err
	}

	// Transfer top-level metadata from DB model to ruleenginepb structure
	if rb != nil {
		rb.Id = dbRB.GetId()
		rb.Name = dbRB.GetName()
		rb.Description = dbRB.GetDescription()
	}

	return &securitypb.GetRuleBasedResponse{RuleBased: rb}, nil
}

//nolint:funlen
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

	dbRB.Name = req.GetRuleBased().GetName()
	dbRB.Description = req.GetRuleBased().GetDescription()

	if req.GetRuleBased().GetNode() != nil {
		b, err := protojson.MarshalOptions{EmitUnpopulated: true}.
			Marshal(req.GetRuleBased().GetNode())
		if err != nil {
			return nil, err
		}
		dbRB.Node = string(b)
	} else {
		dbRB.Node = ""
	}

	if err := srv.ruleBasedRepo.Update(ctx, dbRB); err != nil {
		return nil, err
	}

	rb, err := srv.mapDBToRuleBased(ctx, dbRB)
	if err != nil {
		return nil, err
	}

	if rb != nil {
		rb.Id = dbRB.GetId()
		rb.Name = dbRB.GetName()
		rb.Description = dbRB.GetDescription()
	}

	return &securitypb.UpdateRuleBasedResponse{RuleBased: rb}, nil
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

	var rbs []*ruleenginepb.RuleBased
	for _, dbRB := range dbRBs {
		rb, err := srv.mapDBToRuleBased(ctx, dbRB)
		if err != nil {
			return nil, err
		}
		if rb != nil {
			rb.Id = dbRB.GetId()
			rb.Name = dbRB.GetName()
			rb.Description = dbRB.GetDescription()
			rbs = append(rbs, rb)
		}
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
