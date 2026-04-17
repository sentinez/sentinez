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

	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/security/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	"github.com/sentinez/sentinez/pkg/common/errorx"
)

var _ securitypb.SecurityServiceServer = (*SecurityService)(nil)

// New creates a new SecurityService.
func New(config *confpb.Config) *SecurityService {
	return &SecurityService{
		config: config,
	}
}

// SecurityService handles security operations.
type SecurityService struct {
	config *confpb.Config
}

// ── Expr ──────────────────────────────────────────────────────────────────

func (srv *SecurityService) CreateExpr(ctx context.Context, req *securitypb.CreateExprRequest) (*securitypb.CreateExprResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

func (srv *SecurityService) GetExpr(ctx context.Context, req *securitypb.GetExprRequest) (*securitypb.GetExprResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

func (srv *SecurityService) UpdateExpr(ctx context.Context, req *securitypb.UpdateExprRequest) (*securitypb.UpdateExprResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

func (srv *SecurityService) DeleteExpr(ctx context.Context, req *securitypb.DeleteExprRequest) (*securitypb.DeleteExprResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

func (srv *SecurityService) ListExprs(ctx context.Context, req *securitypb.ListExprsRequest) (*securitypb.ListExprsResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

// ── Rule ──────────────────────────────────────────────────────────────────

func (srv *SecurityService) CreateRule(ctx context.Context, req *securitypb.CreateRuleRequest) (*securitypb.CreateRuleResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

func (srv *SecurityService) GetRule(ctx context.Context, req *securitypb.GetRuleRequest) (*securitypb.GetRuleResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

func (srv *SecurityService) UpdateRule(ctx context.Context, req *securitypb.UpdateRuleRequest) (*securitypb.UpdateRuleResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

func (srv *SecurityService) DeleteRule(ctx context.Context, req *securitypb.DeleteRuleRequest) (*securitypb.DeleteRuleResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

func (srv *SecurityService) ListRules(ctx context.Context, req *securitypb.ListRulesRequest) (*securitypb.ListRulesResponse, error) {
	return nil, errorx.StatusUnimplementedF("unimplemented")
}

// ── Status ────────────────────────────────────────────────────────────────

func (srv *SecurityService) Status(ctx context.Context, req *securitypb.StatusRequest) (*securitypb.StatusResponse, error) {
	return &securitypb.StatusResponse{Msg: "OK"}, nil
}
