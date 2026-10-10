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

	securitysvc "github.com/sentinez/controlplane/security/v1/service"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
)

var _ securitypb.SecurityServiceServer = (*Security)(nil)

// New creates a new Security handler.
func New(
	service *securitysvc.SecurityService,
) securitypb.SecurityServiceServer {
	return &Security{
		service: service,
	}
}

// Security implements the SecurityServiceServer interface.
type Security struct {
	service *securitysvc.SecurityService
}

func (h *Security) CreateSecRule(ctx context.Context,
	req *securitypb.CreateSecRuleRequest,
) (*securitypb.CreateSecRuleResponse, error) {
	return h.service.CreateSecRule(ctx, req)
}

func (h *Security) GetSecRule(ctx context.Context,
	req *securitypb.GetSecRuleRequest,
) (*securitypb.GetSecRuleResponse, error) {
	return h.service.GetSecRule(ctx, req)
}

func (h *Security) UpdateSecRule(ctx context.Context,
	req *securitypb.UpdateSecRuleRequest,
) (*securitypb.UpdateSecRuleResponse, error) {
	return h.service.UpdateSecRule(ctx, req)
}

func (h *Security) DeleteSecRule(ctx context.Context,
	req *securitypb.DeleteSecRuleRequest,
) (*securitypb.DeleteSecRuleResponse, error) {
	return h.service.DeleteSecRule(ctx, req)
}

func (h *Security) ListSecRules(ctx context.Context,
	req *securitypb.ListSecRulesRequest,
) (*securitypb.ListSecRulesResponse, error) {
	return h.service.ListSecRules(ctx, req)
}

func (h *Security) CreateRateLimit(ctx context.Context,
	req *securitypb.CreateRateLimitRequest,
) (*securitypb.CreateRateLimitResponse, error) {
	return h.service.CreateRateLimit(ctx, req)
}

func (h *Security) GetRateLimit(ctx context.Context,
	req *securitypb.GetRateLimitRequest,
) (*securitypb.GetRateLimitResponse, error) {
	return h.service.GetRateLimit(ctx, req)
}

func (h *Security) UpdateRateLimit(ctx context.Context,
	req *securitypb.UpdateRateLimitRequest,
) (*securitypb.UpdateRateLimitResponse, error) {
	return h.service.UpdateRateLimit(ctx, req)
}

func (h *Security) DeleteRateLimit(ctx context.Context,
	req *securitypb.DeleteRateLimitRequest,
) (*securitypb.DeleteRateLimitResponse, error) {
	return h.service.DeleteRateLimit(ctx, req)
}

func (h *Security) ListRateLimits(ctx context.Context,
	req *securitypb.ListRateLimitsRequest,
) (*securitypb.ListRateLimitsResponse, error) {
	return h.service.ListRateLimits(ctx, req)
}

func (h *Security) CreateCoreRuleset(ctx context.Context,
	req *securitypb.CreateCoreRulesetRequest,
) (*securitypb.CreateCoreRulesetResponse, error) {
	return h.service.CreateCoreRuleset(ctx, req)
}

func (h *Security) GetCoreRuleset(ctx context.Context,
	req *securitypb.GetCoreRulesetRequest,
) (*securitypb.GetCoreRulesetResponse, error) {
	return h.service.GetCoreRuleset(ctx, req)
}

func (h *Security) UpdateCoreRuleset(ctx context.Context,
	req *securitypb.UpdateCoreRulesetRequest,
) (*securitypb.UpdateCoreRulesetResponse, error) {
	return h.service.UpdateCoreRuleset(ctx, req)
}

func (h *Security) DeleteCoreRuleset(ctx context.Context,
	req *securitypb.DeleteCoreRulesetRequest,
) (*securitypb.DeleteCoreRulesetResponse, error) {
	return h.service.DeleteCoreRuleset(ctx, req)
}

func (h *Security) ListCoreRulesets(ctx context.Context,
	req *securitypb.ListCoreRulesetsRequest,
) (*securitypb.ListCoreRulesetsResponse, error) {
	return h.service.ListCoreRulesets(ctx, req)
}

func (h *Security) GetRuleInfo(ctx context.Context,
	req *securitypb.GetRuleInfoRequest,
) (*securitypb.GetRuleInfoResponse, error) {
	return h.service.GetRuleInfo(ctx, req)
}

func (h *Security) ListRuleInfos(ctx context.Context,
	req *securitypb.ListRuleInfosRequest,
) (*securitypb.ListRuleInfosResponse, error) {
	return h.service.ListRuleInfos(ctx, req)
}

func (h *Security) Status(ctx context.Context,
	req *securitypb.StatusRequest,
) (*securitypb.StatusResponse, error) {
	return h.service.Status(ctx, req)
}
