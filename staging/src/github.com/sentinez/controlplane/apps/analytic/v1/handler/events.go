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

	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
)

// ListHttpIngressEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListHttpIngressEvents(ctx context.Context,
	req *pb.ListHttpIngressEventsRequest,
) (*pb.ListHttpIngressEventsResponse, error) {
	return hdl.service.ListHttpIngressEvents(ctx, req)
}

// ListHttpEgressEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListHttpEgressEvents(ctx context.Context,
	req *pb.ListHttpEgressEventsRequest,
) (*pb.ListHttpEgressEventsResponse, error) {
	return hdl.service.ListHttpEgressEvents(ctx, req)
}

// GetEgressLatency implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetEgressLatency(ctx context.Context,
	req *pb.GetEgressLatencyRequest,
) (*pb.GetEgressLatencyResponse, error) {
	return hdl.service.GetEgressLatency(ctx, req)
}

// ListSecRuleEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListSecRuleEvents(ctx context.Context,
	req *pb.ListSecRuleEventsRequest,
) (*pb.ListSecRuleEventsResponse, error) {
	return hdl.service.ListSecRuleEvents(ctx, req)
}

// ListRateLimitEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListRateLimitEvents(ctx context.Context,
	req *pb.ListRateLimitEventsRequest,
) (*pb.ListRateLimitEventsResponse, error) {
	return hdl.service.ListRateLimitEvents(ctx, req)
}

// ListCoreRuleEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListCoreRuleEvents(ctx context.Context,
	req *pb.ListCoreRuleEventsRequest,
) (*pb.ListCoreRuleEventsResponse, error) {
	return hdl.service.ListCoreRuleEvents(ctx, req)
}

// GetEventHistogram implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetEventHistogram(ctx context.Context,
	req *pb.GetEventHistogramRequest,
) (*pb.GetEventHistogramResponse, error) {
	return hdl.service.GetEventHistogram(ctx, req)
}

// GetUniqueVisitorHistogram implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetUniqueVisitorHistogram(ctx context.Context,
	req *pb.GetUniqueVisitorHistogramRequest,
) (*pb.GetUniqueVisitorHistogramResponse, error) {
	return hdl.service.GetUniqueVisitorHistogram(ctx, req)
}

// ListVisitors implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListVisitors(ctx context.Context,
	req *pb.ListVisitorsRequest,
) (*pb.ListVisitorsResponse, error) {
	return hdl.service.ListVisitors(ctx, req)
}

// GetEventTopValues implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetEventTopValues(ctx context.Context,
	req *pb.GetEventTopValuesRequest,
) (*pb.GetEventTopValuesResponse, error) {
	return hdl.service.GetEventTopValues(ctx, req)
}
