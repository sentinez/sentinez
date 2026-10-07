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

func (h *Security) Status(ctx context.Context,
	req *securitypb.StatusRequest,
) (*securitypb.StatusResponse, error) {
	return h.service.Status(ctx, req)
}
