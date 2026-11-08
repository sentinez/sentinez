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
	"errors"
	"strings"
	"time"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListHttpIngressEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListHttpIngressEvents(ctx context.Context,
	req *pb.ListHttpIngressEventsRequest) (*pb.ListHttpIngressEventsResponse, error) {
	events, err := listEvents(ctx, req, srv.events.HttpIngress.List)
	if err != nil {
		return nil, err
	}

	return &pb.ListHttpIngressEventsResponse{Events: events}, nil
}

// ListHttpEgressEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListHttpEgressEvents(ctx context.Context,
	req *pb.ListHttpEgressEventsRequest) (*pb.ListHttpEgressEventsResponse, error) {
	events, err := listEvents(ctx, req, srv.events.HttpEgress.List)
	if err != nil {
		return nil, err
	}

	return &pb.ListHttpEgressEventsResponse{Events: events}, nil
}

// ListSecRuleEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListSecRuleEvents(ctx context.Context,
	req *pb.ListSecRuleEventsRequest) (*pb.ListSecRuleEventsResponse, error) {
	events, err := listEvents(ctx, req, srv.events.SecRule.List)
	if err != nil {
		return nil, err
	}

	return &pb.ListSecRuleEventsResponse{Events: events}, nil
}

// ListRateLimitEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListRateLimitEvents(ctx context.Context,
	req *pb.ListRateLimitEventsRequest) (*pb.ListRateLimitEventsResponse, error) {
	events, err := listEvents(ctx, req, srv.events.RateLimit.List)
	if err != nil {
		return nil, err
	}

	return &pb.ListRateLimitEventsResponse{Events: events}, nil
}

// ListCoreRuleEvents implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListCoreRuleEvents(ctx context.Context,
	req *pb.ListCoreRuleEventsRequest) (*pb.ListCoreRuleEventsResponse, error) {
	events, err := listEvents(ctx, req, srv.events.CoreRule.List)
	if err != nil {
		return nil, err
	}

	return &pb.ListCoreRuleEventsResponse{Events: events}, nil
}

// GetEventHistogram implements pb.AnalyticServiceServer.
func (srv *AnalyticService) GetEventHistogram(ctx context.Context,
	req *pb.GetEventHistogramRequest) (*pb.GetEventHistogramResponse, error) {
	points, err := srv.events.Histogram(ctx, req.GetKind(),
		histogramFilter(req))
	if err != nil {
		return nil, err
	}

	return &pb.GetEventHistogramResponse{Points: histogramPoints(points)},
		nil
}

// GetUniqueVisitorHistogram implements pb.AnalyticServiceServer.
func (srv *AnalyticService) GetUniqueVisitorHistogram(ctx context.Context,
	req *pb.GetUniqueVisitorHistogramRequest,
) (*pb.GetUniqueVisitorHistogramResponse, error) {
	points, err := srv.events.HttpIngress.UniqueVisitors(ctx,
		histogramFilter(req))
	if err != nil {
		return nil, err
	}

	return &pb.GetUniqueVisitorHistogramResponse{
		Points: histogramPoints(points),
	}, nil
}

// ListVisitors implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListVisitors(ctx context.Context,
	req *pb.ListVisitorsRequest) (*pb.ListVisitorsResponse, error) {
	points, err := srv.events.HttpIngress.Visitors(ctx, histogramFilter(req))
	if err != nil {
		return nil, err
	}

	resp := &pb.ListVisitorsResponse{
		Visitors: make([]*pb.VisitorPoint, 0, len(points)),
	}
	for _, point := range points {
		resp.Visitors = append(resp.Visitors, &pb.VisitorPoint{
			Time:       timestamppb.New(point.Time),
			ResourceId: point.ResourceId,
			ClientIp:   point.Value,
			Total:      point.Total,
		})
	}

	return resp, nil
}

// histogramRequest the filters every histogram has.
type histogramRequest interface {
	GetResolution() pb.HistogramResolution
	GetResourceIds() []string
	GetFrom() *timestamppb.Timestamp
	GetTo() *timestamppb.Timestamp
}

// histogramFilter the filter of the histogram of req.
func histogramFilter(req histogramRequest) *histogram.Filter {
	return &histogram.Filter{
		Resolution:  histogram.Resolution(req.GetResolution()),
		ResourceIds: req.GetResourceIds(),
		From:        asTime(req.GetFrom()),
		To:          asTime(req.GetTo()),
	}
}

// histogramPoints the proto of points.
func histogramPoints(points []*histogram.Point) []*pb.HistogramPoint {
	resp := make([]*pb.HistogramPoint, 0, len(points))
	for _, point := range points {
		resp = append(resp, &pb.HistogramPoint{
			Time:       timestamppb.New(point.Time),
			ResourceId: point.ResourceId,
			Key:        point.Key,
			Total:      point.Total,
		})
	}

	return resp
}

// The size of a page of events, by default and at most.
const (
	_pageSize    = 50
	_maxPageSize = 500
)

// listRequest the filters every list of events has.
type listRequest interface {
	GetPage() *typepb.Pages
	GetFilters() []string
	GetRequestIds() []string
	GetResourceIds() []string
	GetFrom() *timestamppb.Timestamp
	GetTo() *timestamppb.Timestamp
}

// listEvents the events list returns for the filter of req; an unknown
// field is an invalid argument.
func listEvents[T any](ctx context.Context, req listRequest,
	list func(context.Context, *eventfilter.Filter) ([]*T, error),
) ([]*T, error) {
	filter, err := eventFilter(req)
	if err != nil {
		return nil, err
	}

	events, err := list(ctx, filter)
	if err != nil {
		return nil, fieldError(err)
	}

	return events, nil
}

// eventFilter the filter of the list of req; a filter that is not
// "field:value" is an invalid argument.
func eventFilter(req listRequest) (*eventfilter.Filter, error) {
	matches, err := matchesOf(req.GetFilters())
	if err != nil {
		return nil, err
	}

	return &eventfilter.Filter{
		RequestIds:  req.GetRequestIds(),
		ResourceIds: req.GetResourceIds(),
		Matches:     matches,
		From:        asTime(req.GetFrom()),
		To:          asTime(req.GetTo()),
		Page:        eventPage(req.GetPage()),
	}, nil
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

// matchesOf the eventfilter.Match of filters, each "field:value" (the
// field before the first colon); one without a colon is an invalid
// argument.
func matchesOf(filters []string) ([]eventfilter.Match, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	matches := make([]eventfilter.Match, 0, len(filters))
	for _, f := range filters {
		field, value, ok := strings.Cut(f, ":")
		if !ok || field == "" {
			return nil, szerror.StatusInvalidArgumentF(
				"invalid filter %q: want field:value", f)
		}
		matches = append(matches,
			eventfilter.Match{Field: field, Value: value})
	}

	return matches, nil
}

// fieldError err, an invalid argument when it names a field the events do
// not have (eventfilter.ErrUnknownField).
func fieldError(err error) error {
	if errors.Is(err, eventfilter.ErrUnknownField) {
		return szerror.StatusInvalidArgumentF("%s", err)
	}

	return err
}
