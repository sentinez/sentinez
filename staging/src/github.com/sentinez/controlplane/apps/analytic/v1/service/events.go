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

package analyticsvc

import (
	"context"
	"time"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/histogram"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListHttpIngressEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListHttpIngressEvents(ctx context.Context,
	req *pb.ListHttpIngressEventsRequest) (*pb.ListHttpIngressEventsResponse, error) {
	events, err := srv.events.HttpIngress.List(ctx, eventFilter(req))
	if err != nil {
		return nil, err
	}

	return &pb.ListHttpIngressEventsResponse{Events: events}, nil
}

// ListHttpEgressEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListHttpEgressEvents(ctx context.Context,
	req *pb.ListHttpEgressEventsRequest) (*pb.ListHttpEgressEventsResponse, error) {
	events, err := srv.events.HttpEgress.List(ctx, eventFilter(req))
	if err != nil {
		return nil, err
	}

	return &pb.ListHttpEgressEventsResponse{Events: events}, nil
}

// ListSecRuleEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListSecRuleEvents(ctx context.Context,
	req *pb.ListSecRuleEventsRequest) (*pb.ListSecRuleEventsResponse, error) {
	events, err := srv.events.SecRule.List(ctx, eventFilter(req))
	if err != nil {
		return nil, err
	}

	return &pb.ListSecRuleEventsResponse{Events: events}, nil
}

// ListRateLimitEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListRateLimitEvents(ctx context.Context,
	req *pb.ListRateLimitEventsRequest) (*pb.ListRateLimitEventsResponse, error) {
	events, err := srv.events.RateLimit.List(ctx, eventFilter(req))
	if err != nil {
		return nil, err
	}

	return &pb.ListRateLimitEventsResponse{Events: events}, nil
}

// ListCoreRuleEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListCoreRuleEvents(ctx context.Context,
	req *pb.ListCoreRuleEventsRequest) (*pb.ListCoreRuleEventsResponse, error) {
	events, err := srv.events.CoreRule.List(ctx, eventFilter(req))
	if err != nil {
		return nil, err
	}

	return &pb.ListCoreRuleEventsResponse{Events: events}, nil
}

// GetEventHistogram implements pb.AnalyticServiceServer.
func (srv *AnalyticService) GetEventHistogram(ctx context.Context,
	req *pb.GetEventHistogramRequest) (*pb.GetEventHistogramResponse, error) {
	filter := &histogram.Filter{
		Resolution:  histogram.Resolution(req.GetResolution()),
		ResourceIds: req.GetResourceIds(),
		From:        asTime(req.GetFrom()),
		To:          asTime(req.GetTo()),
	}

	points, err := srv.events.Histogram(ctx, req.GetKind(), filter)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetEventHistogramResponse{
		Points: make([]*pb.HistogramPoint, 0, len(points)),
	}
	for _, point := range points {
		resp.Points = append(resp.Points, &pb.HistogramPoint{
			Time:       timestamppb.New(point.Time),
			ResourceId: point.ResourceId,
			Key:        point.Key,
			Total:      point.Total,
		})
	}

	return resp, nil
}

// The size of a page of events, by default and at most.
const (
	_pageSize    = 50
	_maxPageSize = 500
)

// listRequest the filters every list of events has.
type listRequest interface {
	GetPage() *typepb.Pages
	GetResourceIds() []string
	GetFrom() *timestamppb.Timestamp
	GetTo() *timestamppb.Timestamp
}

// eventFilter the filter of the list of req.
func eventFilter(req listRequest) *eventfilter.Filter {
	return &eventfilter.Filter{
		ResourceIds: req.GetResourceIds(),
		From:        asTime(req.GetFrom()),
		To:          asTime(req.GetTo()),
		Page:        eventPage(req.GetPage()),
	}
}

// eventPage page bounded to _maxPageSize events, the first _pageSize
// ones when unset: a list never reads a whole event table.
func eventPage(page *typepb.Pages) *typepb.Pages {
	size, index := page.GetSize(), page.GetIndex()
	if size <= 0 {
		size = _pageSize
	}
	if index <= 0 {
		index = 1
	}

	return &typepb.Pages{Size: min(size, _maxPageSize), Index: index}
}

// asTime the time of ts, zero when unset.
func asTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}

	return ts.AsTime()
}
