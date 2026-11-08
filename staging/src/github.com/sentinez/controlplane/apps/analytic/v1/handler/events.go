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

	"github.com/sentinez/controlplane/pkg/headers"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
)

// ListHttpIngressEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListHttpIngressEvents(ctx context.Context,
	req *pb.ListHttpIngressEventsRequest,
) (*pb.ListHttpIngressEventsResponse, error) {
	err := headers.Allow(ctx, pb.GetAnalyticServiceListHttpIngressEvents())
	if err != nil {
		return nil, err
	}

	return hdl.service.ListHttpIngressEvents(ctx, req)
}

// ListHttpEgressEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListHttpEgressEvents(ctx context.Context,
	req *pb.ListHttpEgressEventsRequest,
) (*pb.ListHttpEgressEventsResponse, error) {
	err := headers.Allow(ctx, pb.GetAnalyticServiceListHttpEgressEvents())
	if err != nil {
		return nil, err
	}

	return hdl.service.ListHttpEgressEvents(ctx, req)
}

// ListSecRuleEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListSecRuleEvents(ctx context.Context,
	req *pb.ListSecRuleEventsRequest,
) (*pb.ListSecRuleEventsResponse, error) {
	err := headers.Allow(ctx, pb.GetAnalyticServiceListSecRuleEvents())
	if err != nil {
		return nil, err
	}

	return hdl.service.ListSecRuleEvents(ctx, req)
}

// ListRateLimitEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListRateLimitEvents(ctx context.Context,
	req *pb.ListRateLimitEventsRequest,
) (*pb.ListRateLimitEventsResponse, error) {
	err := headers.Allow(ctx, pb.GetAnalyticServiceListRateLimitEvents())
	if err != nil {
		return nil, err
	}

	return hdl.service.ListRateLimitEvents(ctx, req)
}

// ListCoreRuleEvents implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListCoreRuleEvents(ctx context.Context,
	req *pb.ListCoreRuleEventsRequest,
) (*pb.ListCoreRuleEventsResponse, error) {
	err := headers.Allow(ctx, pb.GetAnalyticServiceListCoreRuleEvents())
	if err != nil {
		return nil, err
	}

	return hdl.service.ListCoreRuleEvents(ctx, req)
}

// GetEventHistogram implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetEventHistogram(ctx context.Context,
	req *pb.GetEventHistogramRequest,
) (*pb.GetEventHistogramResponse, error) {
	err := headers.Allow(ctx, pb.GetAnalyticServiceGetEventHistogram())
	if err != nil {
		return nil, err
	}

	return hdl.service.GetEventHistogram(ctx, req)
}

// GetUniqueVisitorHistogram implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetUniqueVisitorHistogram(ctx context.Context,
	req *pb.GetUniqueVisitorHistogramRequest,
) (*pb.GetUniqueVisitorHistogramResponse, error) {
	err := headers.Allow(ctx,
		pb.GetAnalyticServiceGetUniqueVisitorHistogram())
	if err != nil {
		return nil, err
	}

	return hdl.service.GetUniqueVisitorHistogram(ctx, req)
}

// ListVisitors implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) ListVisitors(ctx context.Context,
	req *pb.ListVisitorsRequest,
) (*pb.ListVisitorsResponse, error) {
	err := headers.Allow(ctx, pb.GetAnalyticServiceListVisitors())
	if err != nil {
		return nil, err
	}

	return hdl.service.ListVisitors(ctx, req)
}

// GetEventTopValues implements pb.AnalyticServiceServer.
func (hdl *AnalyticHandler) GetEventTopValues(ctx context.Context,
	req *pb.GetEventTopValuesRequest,
) (*pb.GetEventTopValuesResponse, error) {
	err := headers.Allow(ctx, pb.GetAnalyticServiceGetEventTopValues())
	if err != nil {
		return nil, err
	}

	return hdl.service.GetEventTopValues(ctx, req)
}
