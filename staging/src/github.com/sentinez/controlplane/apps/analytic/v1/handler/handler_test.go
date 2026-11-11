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

package analytichdl

import (
	"context"
	"testing"

	membermock "github.com/sentinez/controlplane/apps/tenant/v1/repos/members/mock"
	resourcemock "github.com/sentinez/controlplane/apps/tenant/v1/repos/resources/mock"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/controlplane/pkg/headers"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// noService is the handler of a caller who sees no resource: the service
// is nil, so a call that reaches it panics.
func noService(t *testing.T) *AnalyticHandler {
	t.Helper()

	members := membermock.NewMockIMember(t)
	members.EXPECT().List(mock.Anything, mock.Anything).
		Return(&tenantpb.ListMemberResponse{}, nil).Maybe()
	resources := resourcemock.NewMockIResource(t)

	return &AnalyticHandler{authz: authz.New(members, resources)}
}

func portal(t *testing.T) context.Context {
	t.Helper()

	return headers.WithAuth(t.Context(), &headers.Auth{
		Context: &typepb.Context{
			UserId:       "user_1",
			ControlPlane: typepb.ControlPlane_CONTROL_PLANE_PORTAL,
		},
	})
}

func TestCallerWithoutResourcesSeesNothing(t *testing.T) {
	tests := []struct {
		name string
		call func(*AnalyticHandler, context.Context) error
	}{
		{"ingress events", func(h *AnalyticHandler, c context.Context) error {
			_, err := h.ListHttpIngressEvents(c,
				&pb.ListHttpIngressEventsRequest{})
			return err
		}},
		{"egress events", func(h *AnalyticHandler, c context.Context) error {
			_, err := h.ListHttpEgressEvents(c,
				&pb.ListHttpEgressEventsRequest{})
			return err
		}},
		{"sec rule events", func(h *AnalyticHandler, c context.Context) error {
			_, err := h.ListSecRuleEvents(c, &pb.ListSecRuleEventsRequest{})
			return err
		}},
		{"histogram", func(h *AnalyticHandler, c context.Context) error {
			_, err := h.GetEventHistogram(c, &pb.GetEventHistogramRequest{})
			return err
		}},
		{"top values", func(h *AnalyticHandler, c context.Context) error {
			_, err := h.GetEventTopValues(c, &pb.GetEventTopValuesRequest{})
			return err
		}},
		{"visitors", func(h *AnalyticHandler, c context.Context) error {
			_, err := h.ListVisitors(c, &pb.ListVisitorsRequest{})
			return err
		}},
		{"latency", func(h *AnalyticHandler, c context.Context) error {
			_, err := h.GetEgressLatency(c, &pb.GetEgressLatencyRequest{})
			return err
		}},
		{"activities", func(h *AnalyticHandler, c context.Context) error {
			_, err := h.ListActivities(c, &pb.ListActivitiesRequest{})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, tt.call(noService(t), portal(t)))
		})
	}
}

func TestCallWithoutCallerFails(t *testing.T) {
	_, err := noService(t).ListHttpIngressEvents(t.Context(),
		&pb.ListHttpIngressEventsRequest{})

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}
