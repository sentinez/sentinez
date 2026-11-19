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

	securitysvc "github.com/sentinez/controlplane/apps/security/v1/service"
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

func (h *Security) CreateSecRateLimit(ctx context.Context,
	req *securitypb.CreateSecRateLimitRequest,
) (*securitypb.CreateSecRateLimitResponse, error) {
	return h.service.CreateSecRateLimit(ctx, req)
}

func (h *Security) GetSecRateLimit(ctx context.Context,
	req *securitypb.GetSecRateLimitRequest,
) (*securitypb.GetSecRateLimitResponse, error) {
	return h.service.GetSecRateLimit(ctx, req)
}

func (h *Security) UpdateSecRateLimit(ctx context.Context,
	req *securitypb.UpdateSecRateLimitRequest,
) (*securitypb.UpdateSecRateLimitResponse, error) {
	return h.service.UpdateSecRateLimit(ctx, req)
}

func (h *Security) DeleteSecRateLimit(ctx context.Context,
	req *securitypb.DeleteSecRateLimitRequest,
) (*securitypb.DeleteSecRateLimitResponse, error) {
	return h.service.DeleteSecRateLimit(ctx, req)
}

func (h *Security) ListSecRateLimits(ctx context.Context,
	req *securitypb.ListSecRateLimitsRequest,
) (*securitypb.ListSecRateLimitsResponse, error) {
	return h.service.ListSecRateLimits(ctx, req)
}

func (h *Security) CreateSecRuleset(ctx context.Context,
	req *securitypb.CreateSecRulesetRequest,
) (*securitypb.CreateSecRulesetResponse, error) {
	return h.service.CreateSecRuleset(ctx, req)
}

func (h *Security) GetSecRuleset(ctx context.Context,
	req *securitypb.GetSecRulesetRequest,
) (*securitypb.GetSecRulesetResponse, error) {
	return h.service.GetSecRuleset(ctx, req)
}

func (h *Security) UpdateSecRuleset(ctx context.Context,
	req *securitypb.UpdateSecRulesetRequest,
) (*securitypb.UpdateSecRulesetResponse, error) {
	return h.service.UpdateSecRuleset(ctx, req)
}

func (h *Security) DeleteSecRuleset(ctx context.Context,
	req *securitypb.DeleteSecRulesetRequest,
) (*securitypb.DeleteSecRulesetResponse, error) {
	return h.service.DeleteSecRuleset(ctx, req)
}

func (h *Security) ListSecRulesets(ctx context.Context,
	req *securitypb.ListSecRulesetsRequest,
) (*securitypb.ListSecRulesetsResponse, error) {
	return h.service.ListSecRulesets(ctx, req)
}

func (h *Security) GetCRSRuleInfo(ctx context.Context,
	req *securitypb.GetCRSRuleInfoRequest,
) (*securitypb.GetCRSRuleInfoResponse, error) {
	return h.service.GetCRSRuleInfo(ctx, req)
}

func (h *Security) ListCRSRuleInfos(ctx context.Context,
	req *securitypb.ListCRSRuleInfosRequest,
) (*securitypb.ListCRSRuleInfosResponse, error) {
	return h.service.ListCRSRuleInfos(ctx, req)
}

func (h *Security) ListCRSRuleTags(ctx context.Context,
	req *securitypb.ListCRSRuleTagsRequest,
) (*securitypb.ListCRSRuleTagsResponse, error) {
	return h.service.ListCRSRuleTags(ctx, req)
}

func (h *Security) Status(ctx context.Context,
	req *securitypb.StatusRequest,
) (*securitypb.StatusResponse, error) {
	return h.service.Status(ctx, req)
}
