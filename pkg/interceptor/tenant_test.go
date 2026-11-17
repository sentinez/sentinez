// Copyright 2026 Duc-Hung Ho.
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

package interceptor

import (
	"testing"

	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestTenantIdOf(t *testing.T) {
	for _, tt := range _tenantCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tenantIdOf(tt.give...))
		})
	}
}

var _tenantCases = []struct {
	name string
	give []proto.Message
	want string
}{
	{
		name: "tenant id of the request",
		give: []proto.Message{
			&tenantpb.CreateMemberRequest{TenantId: "tnt_1"},
		},
		want: "tnt_1",
	},
	{
		name: "owner nested in a message",
		give: []proto.Message{&tenantpb.CaCertificate{
			Owner: &typepb.Owner{TenantId: "tnt_2"},
		}},
		want: "tnt_2",
	},
	{
		name: "tenant id only in the response",
		give: []proto.Message{
			&tenantpb.CreateMemberRequest{},
			&tenantpb.CaCertificate{
				Owner: &typepb.Owner{TenantId: "tnt_3"},
			},
		},
		want: "tnt_3",
	},
	{
		name: "no tenant id, nil message",
		give: []proto.Message{nil, &tenantpb.CreateMemberRequest{}},
	},
}
