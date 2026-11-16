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

package tenantsvc

import (
	"testing"

	resourcemock "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources/mock"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func maskOf(paths ...string) *fieldmaskpb.FieldMask {
	return &fieldmaskpb.FieldMask{Paths: paths}
}

// storedResource is the resource _resourceId of _tenantId, served at
// shop.example.com with one location.
func storedResource() *tenantpb.Resource {
	return &tenantpb.Resource{
		Id:        _resourceId,
		TenantId:  _tenantId,
		Name:      "shop",
		Domain:    "shop.example.com",
		Locations: []*edgepb.Location{{Location: "/"}},
		Status:    typepb.Status_STATUS_ACTIVE,
	}
}

// resourcesWith is a repo holding storedResource; it expects the update
// of want under paths when want is not nil.
func resourcesWith(t *testing.T, want *tenantpb.Resource, paths []string,
) *resourcemock.MockIResource {
	t.Helper()

	repo := resourcemock.NewMockIResource(t)
	repo.EXPECT().Get(mock.Anything, _resourceId).
		Return(storedResource(), nil)

	if want != nil {
		repo.EXPECT().Update(mock.Anything, want, paths).Return(nil)
	}

	return repo
}

func TestUpdateResourceMask(t *testing.T) {
	want := storedResource()
	want.Locations = nil
	want.Status = typepb.Status_STATUS_DISABLE
	paths := []string{tenantpb.Resource_Locations, tenantpb.Resource_Status}

	svc := New(nil, resourcesWith(t, want, paths), nil, nil, nil)
	_, err := svc.UpdateResource(asAdmin(t), &tenantpb.UpdateResourceRequest{
		Id: _resourceId,
		// the name is not in the mask, the locations are cleared
		Resource: &tenantpb.Resource{
			Name: "other", Status: typepb.Status_STATUS_DISABLE,
		},
		UpdateMask: maskOf(paths...),
	})

	require.NoError(t, err)
}

func TestUpdateResourceImpliedMask(t *testing.T) {
	want := storedResource()
	want.Domain = "new.example.com"

	svc := New(nil, resourcesWith(t, want,
		[]string{tenantpb.Resource_Domain}), nil, nil, nil)
	_, err := svc.UpdateResource(asAdmin(t), &tenantpb.UpdateResourceRequest{
		Id: _resourceId,
		// the edge matches the domain in lower case
		Resource: &tenantpb.Resource{Domain: "New.Example.COM"},
	})

	require.NoError(t, err)
}

func TestUpdateResourceInvalidMask(t *testing.T) {
	tests := []struct {
		name      string
		givePaths []string
	}{
		{"field not updatable", []string{tenantpb.Resource_TenantId}},
		{"domain cleared", []string{tenantpb.Resource_Domain}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(nil, resourcesWith(t, nil, nil), nil, nil, nil)

			_, err := svc.UpdateResource(asAdmin(t),
				&tenantpb.UpdateResourceRequest{
					Id:         _resourceId,
					Resource:   &tenantpb.Resource{TenantId: "tnt_other"},
					UpdateMask: maskOf(tt.givePaths...),
				})

			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}
