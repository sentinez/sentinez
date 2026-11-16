// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package securitysvc

import (
	"testing"

	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func maskOf(paths ...string) *fieldmaskpb.FieldMask {
	return &fieldmaskpb.FieldMask{Paths: paths}
}

func TestUpdateSecRuleWritesTheMask(t *testing.T) {
	f := newFixture(t, typepb.Role_ROLE_LEADER)
	f.rules.byId[_ruleId].Priority = 7

	resp, err := f.svc.UpdateSecRule(portalUser(t),
		&securitypb.UpdateSecRuleRequest{
			Id:         _ruleId,
			SecRule:    &securitypb.SecRule{Name: "block", Priority: 1},
			UpdateMask: maskOf(securitypb.SecRule_Name),
		})

	require.NoError(t, err)
	assert.Equal(t, [][]string{{securitypb.SecRule_Name}}, f.rules.updated)
	assert.Equal(t, "block", resp.GetSecRule().GetName())
	assert.Equal(t, int32(7), resp.GetSecRule().GetPriority())
	assert.Equal(t, _resourceId,
		resp.GetSecRule().GetOwner().GetResourceId())
}

func TestUpdateSecRuleOwnerNotUpdatable(t *testing.T) {
	f := newFixture(t, typepb.Role_ROLE_LEADER)

	_, err := f.svc.UpdateSecRule(portalUser(t),
		&securitypb.UpdateSecRuleRequest{
			Id:         _ruleId,
			SecRule:    &securitypb.SecRule{Owner: ownedBy(_foreignId)},
			UpdateMask: maskOf("owner"),
		})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Empty(t, f.rules.updated)
}

func TestUpdateRateLimitValidatesTheResult(t *testing.T) {
	f := newFixture(t, typepb.Role_ROLE_LEADER)
	stored := f.limits.byId[_ruleId]
	stored.MaxRequests, stored.TimeWindow = 10, "1m"

	// max_requests cleared: the rule left has none
	_, err := f.svc.UpdateRateLimit(portalUser(t),
		&securitypb.UpdateRateLimitRequest{
			Id:         _ruleId,
			RateLimit:  &securitypb.RateLimit{Name: "api"},
			UpdateMask: maskOf(securitypb.RateLimit_MaxRequests),
		})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Empty(t, f.limits.updated)
}
