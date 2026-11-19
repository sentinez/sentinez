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

	"github.com/sentinez/controlplane/pkg/authz"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
)

// ListHttpIngressEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListHttpIngressEvents(ctx context.Context,
	req *pb.ListHttpIngressEventsRequest,
) (*pb.ListHttpIngressEventsResponse, error) {
	req, ok, err := authz.Scoped(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.ListHttpIngressEventsResponse{}, err
	}

	return hdl.service.ListHttpIngressEvents(ctx, req)
}

// ListHttpEgressEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListHttpEgressEvents(ctx context.Context,
	req *pb.ListHttpEgressEventsRequest,
) (*pb.ListHttpEgressEventsResponse, error) {
	req, ok, err := authz.Scoped(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.ListHttpEgressEventsResponse{}, err
	}

	return hdl.service.ListHttpEgressEvents(ctx, req)
}

// GetEgressLatency implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetEgressLatency(ctx context.Context,
	req *pb.GetEgressLatencyRequest,
) (*pb.GetEgressLatencyResponse, error) {
	selection := req.GetSelection()
	if selection == nil {
		selection = &pb.EventSelection{}
	}

	selection, ok, err := authz.Scoped(ctx, hdl.authz, selection)
	if err != nil || !ok {
		return &pb.GetEgressLatencyResponse{}, err
	}

	return hdl.service.GetEgressLatency(ctx,
		&pb.GetEgressLatencyRequest{Selection: selection})
}

// ListSecRuleEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListSecRuleEvents(ctx context.Context,
	req *pb.ListSecRuleEventsRequest,
) (*pb.ListSecRuleEventsResponse, error) {
	req, ok, err := authz.Scoped(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.ListSecRuleEventsResponse{}, err
	}

	return hdl.service.ListSecRuleEvents(ctx, req)
}

// ListSecRateLimitEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListSecRateLimitEvents(ctx context.Context,
	req *pb.ListSecRateLimitEventsRequest,
) (*pb.ListSecRateLimitEventsResponse, error) {
	req, ok, err := authz.Scoped(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.ListSecRateLimitEventsResponse{}, err
	}

	return hdl.service.ListSecRateLimitEvents(ctx, req)
}

// ListSecRulesetEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListSecRulesetEvents(ctx context.Context,
	req *pb.ListSecRulesetEventsRequest,
) (*pb.ListSecRulesetEventsResponse, error) {
	req, ok, err := authz.Scoped(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.ListSecRulesetEventsResponse{}, err
	}

	return hdl.service.ListSecRulesetEvents(ctx, req)
}

// GetEventHistogram implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetEventHistogram(ctx context.Context,
	req *pb.GetEventHistogramRequest,
) (*pb.GetEventHistogramResponse, error) {
	req, ok, err := authz.Scoped(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.GetEventHistogramResponse{}, err
	}

	return hdl.service.GetEventHistogram(ctx, req)
}

// GetUniqueVisitorHistogram implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetUniqueVisitorHistogram(ctx context.Context,
	req *pb.GetUniqueVisitorHistogramRequest,
) (*pb.GetUniqueVisitorHistogramResponse, error) {
	req, ok, err := authz.Scoped(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.GetUniqueVisitorHistogramResponse{}, err
	}

	return hdl.service.GetUniqueVisitorHistogram(ctx, req)
}

// ListVisitors implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListVisitors(ctx context.Context,
	req *pb.ListVisitorsRequest,
) (*pb.ListVisitorsResponse, error) {
	req, ok, err := authz.Scoped(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.ListVisitorsResponse{}, err
	}

	return hdl.service.ListVisitors(ctx, req)
}

// GetEventTopValues implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetEventTopValues(ctx context.Context,
	req *pb.GetEventTopValuesRequest,
) (*pb.GetEventTopValuesResponse, error) {
	req, ok, err := authz.Scoped(ctx, hdl.authz, req)
	if err != nil || !ok {
		return &pb.GetEventTopValuesResponse{}, err
	}

	return hdl.service.GetEventTopValues(ctx, req)
}
