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
package interceptor

import (
	"testing"

	"github.com/sentinez/controlplane/pkg/auditowner"
	secpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

var _ownerCases = []struct {
	name           string
	give           []proto.Message
	wantTenantId   string
	wantResourceId string
}{
	{
		name: "owner nested in the request",
		give: []proto.Message{&secpb.CreateSecRuleRequest{
			SecRule: &secpb.SecRule{Owner: &typepb.Owner{
				TenantId: "tnt_1", ResourceId: "res_1",
			}},
		}},
		wantTenantId: "tnt_1", wantResourceId: "res_1",
	},
	{
		name: "id of the request",
		give: []proto.Message{
			&tenantpb.DeleteResourceRequest{Id: "res_2"},
		},
		wantResourceId: "res_2",
	},
	{
		name: "request first, then the response",
		give: []proto.Message{
			&tenantpb.CreateCaCertificateRequest{ResourceId: "res_3"},
			&tenantpb.CreateCaCertificateResponse{
				CaCertificates: []*tenantpb.CaCertificate{{
					Owner: &typepb.Owner{
						TenantId: "tnt_3", ResourceId: "res_other",
					},
				}},
			},
		},
		wantTenantId: "tnt_3", wantResourceId: "res_3",
	},
	{
		name: "unmarked fields, nil response",
		give: []proto.Message{&tenantpb.DeleteMemberRequest{Id: "mem_1"}, nil},
	},
}

func TestSetOwner(t *testing.T) {
	for _, tt := range _ownerCases {
		t.Run(tt.name, func(t *testing.T) {
			var o auditowner.Owner
			setOwner(&o, tt.give...)
			tenantId, resourceId := o.Ids()
			assert.Equal(t, tt.wantTenantId, tenantId)
			assert.Equal(t, tt.wantResourceId, resourceId)
		})
	}
}
