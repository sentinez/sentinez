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

	"github.com/sentinez/controlplane/apps/security/v1/repos/coreruleset"
	"github.com/sentinez/controlplane/apps/security/v1/repos/ratelimit"
	"github.com/sentinez/controlplane/apps/security/v1/repos/secrule"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/pkg/szerror"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

var _ securitypb.SecurityServiceServer = (*SecurityService)(nil)

// New creates a new SecurityService.
func New(
	config *settingpb.Config,
	secRuleRepo secrule.ISecRule,
	rateLimitRepo ratelimit.IRateLimit,
	coreRulesetRepo coreruleset.ICoreRuleset,
	az *authz.Authorizer,
) *SecurityService {
	return &SecurityService{
		config:          config,
		secRuleRepo:     secRuleRepo,
		rateLimitRepo:   rateLimitRepo,
		coreRulesetRepo: coreRulesetRepo,
		authz:           az,
	}
}

// SecurityService handles security operations.
type SecurityService struct {
	config          *settingpb.Config
	secRuleRepo     secrule.ISecRule
	rateLimitRepo   ratelimit.IRateLimit
	coreRulesetRepo coreruleset.ICoreRuleset
	authz           *authz.Authorizer
}

// newOwner checks the caller may write the resource of owner and returns
// the owner with the tenant of that resource: the tenant of the request is
// not trusted.
func (srv *SecurityService) newOwner(ctx context.Context,
	owner *typepb.Owner) (*typepb.Owner, error) {

	resource, err := srv.authz.
		RequireResource(ctx, owner.GetResourceId(), authz.Write)
	if err != nil {
		return nil, err
	}

	return &typepb.Owner{
		ResourceId: resource.GetId(),
		TenantId:   resource.GetTenantId(),
	}, nil
}

// ownedRow is a rule that belongs to a resource.
type ownedRow interface {
	GetOwner() *typepb.Owner
}

// owned reads the row id with get, NotFound if there is none, and checks
// the caller has at least role least in the tenant of its resource.
func owned[M ownedRow](ctx context.Context, srv *SecurityService,
	get func(context.Context, string) (M, error), what, id string,
	least typepb.Role) (M, error) {

	var none M

	row, err := get(ctx, id)
	if szerror.IsNoRows(err) {
		return none, szerror.StatusNotFoundF("%s not found: id=%s", what, id)
	}
	if err != nil {
		return none, err
	}

	if err = srv.authz.RequireOwner(ctx, row.GetOwner(), least); err != nil {
		return none, err
	}

	return row, nil
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
		Owner:       dbRB.GetOwner(),
	}
}

// ── SecRule ────────────────────────────────────────────────────────────

// CreateSecRule creates a new WAF rule based
func (srv *SecurityService) CreateSecRule(ctx context.Context,
	req *securitypb.CreateSecRuleRequest,
) (*securitypb.CreateSecRuleResponse, error) {
	if req.GetSecRule() == nil {
		return nil, szerror.StatusInvalidArgumentF("sec_rule is required")
	}

	owner, err := srv.newOwner(ctx, req.GetSecRule().GetOwner())
	if err != nil {
		return nil, err
	}

	req.GetSecRule().Owner = owner

	created, err := srv.secRuleRepo.Create(ctx, req.GetSecRule())
	if err != nil {
		return nil, err
	}

	return &securitypb.CreateSecRuleResponse{Id: created.GetId()}, nil
}

func (srv *SecurityService) GetSecRule(ctx context.Context,
	req *securitypb.GetSecRuleRequest,
) (*securitypb.GetSecRuleResponse, error) {
	dbRB, err := owned(ctx, srv, srv.secRuleRepo.Get, "sec rule",
		req.GetId(), authz.Read)
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
		return nil, szerror.StatusInvalidArgumentF("sec_rule is required")
	}

	dbRB, err := owned(ctx, srv, srv.secRuleRepo.Get, "sec rule",
		req.GetId(), authz.Write)
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
	_, err := owned(ctx, srv, srv.secRuleRepo.Get, "sec rule",
		req.GetId(), authz.Write)
	if err != nil {
		return nil, err
	}

	if err = srv.secRuleRepo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}

	return &securitypb.DeleteSecRuleResponse{}, nil
}

func (srv *SecurityService) ListSecRules(ctx context.Context,
	req *securitypb.ListSecRulesRequest,
) (*securitypb.ListSecRulesResponse, error) {
	req, ok, err := authz.Scoped(ctx, srv.authz, req)
	if err != nil || !ok {
		return &securitypb.ListSecRulesResponse{}, err
	}

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
