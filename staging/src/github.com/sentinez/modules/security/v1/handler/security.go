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

	securitysvc "github.com/sentinez/modules/security/v1/service"
	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/security/v1"
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

func (h *Security) CreateRuleBased(ctx context.Context,
	req *securitypb.CreateRuleBasedRequest,
) (*securitypb.CreateRuleBasedResponse, error) {
	return h.service.CreateRuleBased(ctx, req)
}

func (h *Security) GetRuleBased(ctx context.Context,
	req *securitypb.GetRuleBasedRequest,
) (*securitypb.GetRuleBasedResponse, error) {
	return h.service.GetRuleBased(ctx, req)
}

func (h *Security) UpdateRuleBased(ctx context.Context,
	req *securitypb.UpdateRuleBasedRequest,
) (*securitypb.UpdateRuleBasedResponse, error) {
	return h.service.UpdateRuleBased(ctx, req)
}

func (h *Security) DeleteRuleBased(ctx context.Context,
	req *securitypb.DeleteRuleBasedRequest,
) (*securitypb.DeleteRuleBasedResponse, error) {
	return h.service.DeleteRuleBased(ctx, req)
}

func (h *Security) ListRuleBaseds(ctx context.Context,
	req *securitypb.ListRuleBasedsRequest,
) (*securitypb.ListRuleBasedsResponse, error) {
	return h.service.ListRuleBaseds(ctx, req)
}

func (h *Security) Status(ctx context.Context,
	req *securitypb.StatusRequest,
) (*securitypb.StatusResponse, error) {
	return h.service.Status(ctx, req)
}
