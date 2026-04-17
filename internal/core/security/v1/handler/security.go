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

package securityhdl

import (
	"context"

	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/security/v1"
	securitysvc "github.com/sentinez/sentinez/internal/core/security/v1/service"
)

var _ securitypb.SecurityServiceServer = (*Security)(nil)

// New creates a new Security handler.
func New(service *securitysvc.SecurityService) securitypb.SecurityServiceServer {
	return &Security{
		service: service,
	}
}

// Security implements the SecurityServiceServer interface.
type Security struct {
	service *securitysvc.SecurityService
}

// ── Expr ──────────────────────────────────────────────────────────────────

func (h *Security) CreateExpr(ctx context.Context, req *securitypb.CreateExprRequest) (*securitypb.CreateExprResponse, error) {
	return h.service.CreateExpr(ctx, req)
}

func (h *Security) GetExpr(ctx context.Context, req *securitypb.GetExprRequest) (*securitypb.GetExprResponse, error) {
	return h.service.GetExpr(ctx, req)
}

func (h *Security) UpdateExpr(ctx context.Context, req *securitypb.UpdateExprRequest) (*securitypb.UpdateExprResponse, error) {
	return h.service.UpdateExpr(ctx, req)
}

func (h *Security) DeleteExpr(ctx context.Context, req *securitypb.DeleteExprRequest) (*securitypb.DeleteExprResponse, error) {
	return h.service.DeleteExpr(ctx, req)
}

func (h *Security) ListExprs(ctx context.Context, req *securitypb.ListExprsRequest) (*securitypb.ListExprsResponse, error) {
	return h.service.ListExprs(ctx, req)
}

// ── Rule ──────────────────────────────────────────────────────────────────

func (h *Security) CreateRule(ctx context.Context, req *securitypb.CreateRuleRequest) (*securitypb.CreateRuleResponse, error) {
	return h.service.CreateRule(ctx, req)
}

func (h *Security) GetRule(ctx context.Context, req *securitypb.GetRuleRequest) (*securitypb.GetRuleResponse, error) {
	return h.service.GetRule(ctx, req)
}

func (h *Security) UpdateRule(ctx context.Context, req *securitypb.UpdateRuleRequest) (*securitypb.UpdateRuleResponse, error) {
	return h.service.UpdateRule(ctx, req)
}

func (h *Security) DeleteRule(ctx context.Context, req *securitypb.DeleteRuleRequest) (*securitypb.DeleteRuleResponse, error) {
	return h.service.DeleteRule(ctx, req)
}

func (h *Security) ListRules(ctx context.Context, req *securitypb.ListRulesRequest) (*securitypb.ListRulesResponse, error) {
	return h.service.ListRules(ctx, req)
}

// ── Status ────────────────────────────────────────────────────────────────

func (h *Security) Status(ctx context.Context, req *securitypb.StatusRequest) (*securitypb.StatusResponse, error) {
	return h.service.Status(ctx, req)
}
