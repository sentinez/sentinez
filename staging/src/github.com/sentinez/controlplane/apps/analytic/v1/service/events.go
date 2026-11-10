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
	"time"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/pkg/schema"
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
	events, err := histogramEvents(req)
	if err != nil {
		return nil, err
	}

	filter, window, err := histogramFilter(req, time.Now())
	if err != nil {
		return nil, err
	}

	points, err := srv.events.Histogram(ctx, req.GetKind(), filter, events)
	if err != nil {
		return nil, filterError(err)
	}

	return &pb.GetEventHistogramResponse{
		Points:    histogramPoints(points),
		Interval:  window.Interval,
		From:      timestamppb.New(window.From),
		To:        timestamppb.New(window.To),
		Truncated: window.Truncated,
	}, nil
}

// histogramEvents the events the histogram of sel counts: its requests
// and field filters, nil without any (the views count them all); a filter
// that is not "field:value" is an invalid argument.
func histogramEvents(sel eventfilter.Selection) (*eventfilter.Filter, error) {
	if len(sel.GetRequestIds()) == 0 && len(sel.GetFilters()) == 0 {
		return nil, nil
	}

	matches, err := eventfilter.MatchesOf(sel.GetFilters())
	if err != nil {
		return nil, filterError(err)
	}

	return &eventfilter.Filter{
		RequestIds: sel.GetRequestIds(),
		Matches:    matches,
	}, nil
}

// GetUniqueVisitorHistogram implements pb.AnalyticServiceServer.
func (srv *AnalyticService) GetUniqueVisitorHistogram(ctx context.Context,
	req *pb.GetUniqueVisitorHistogramRequest,
) (*pb.GetUniqueVisitorHistogramResponse, error) {
	filter, window, err := histogramFilter(req, time.Now())
	if err != nil {
		return nil, err
	}

	points, err := srv.events.HttpIngress.UniqueVisitors(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &pb.GetUniqueVisitorHistogramResponse{
		Points:    histogramPoints(points),
		Interval:  window.Interval,
		From:      timestamppb.New(window.From),
		To:        timestamppb.New(window.To),
		Truncated: window.Truncated,
	}, nil
}

// ListVisitors implements pb.AnalyticServiceServer.
func (srv *AnalyticService) ListVisitors(ctx context.Context,
	req *pb.ListVisitorsRequest) (*pb.ListVisitorsResponse, error) {
	filter, window, err := histogramFilter(req, time.Now())
	if err != nil {
		return nil, err
	}

	points, err := srv.events.HttpIngress.Visitors(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &pb.ListVisitorsResponse{
		Visitors:  visitorPoints(points),
		Interval:  window.Interval,
		From:      timestamppb.New(window.From),
		To:        timestamppb.New(window.To),
		Truncated: window.Truncated,
	}, nil
}

// visitorPoints the proto of points, the client ip their Value.
func visitorPoints(points []*histogram.Point) []*pb.VisitorPoint {
	visitors := make([]*pb.VisitorPoint, 0, len(points))
	for _, point := range points {
		visitors = append(visitors, &pb.VisitorPoint{
			Time:       timestamppb.New(point.Time),
			ResourceId: point.ResourceId,
			ClientIp:   point.Value,
			Total:      point.Total,
		})
	}

	return visitors
}

// histogramRequest the filters every histogram has.
type histogramRequest interface {
	GetResolution() typepb.HistogramResolution
	GetResourceIds() []string
	GetFrom() *timestamppb.Timestamp
	GetTo() *timestamppb.Timestamp
}

// histogramFilter the filter of the histogram of req, its resolution and
// range resolved at now, and the window of its buckets; from or to older
// than the events are kept is an invalid argument.
func histogramFilter(req histogramRequest, now time.Time) (
	*histogram.Filter, histogram.Window, error) {
	filter := &histogram.Filter{
		Resolution:  req.GetResolution(),
		ResourceIds: req.GetResourceIds(),
		From:        eventfilter.TimeOf(req.GetFrom()),
		To:          eventfilter.TimeOf(req.GetTo()),
	}

	err := keptSince(now, filter.From, filter.To)
	if err != nil {
		return nil, histogram.Window{}, err
	}

	return filter, filter.Resolve(now), nil
}

// keptSince whether the events at the times, unset ones skipped, are still
// kept at now (schema.EventRetention): an invalid argument when one is
// older, their histogram long gone.
func keptSince(now time.Time, times ...time.Time) error {
	oldest := now.Add(-schema.EventRetention)
	for _, t := range times {
		if !t.IsZero() && t.Before(oldest) {
			return szerror.StatusInvalidArgumentF(
				"%s is before %s: the events are kept %s",
				t.Format(time.RFC3339), oldest.Format(time.RFC3339),
				schema.EventRetention)
		}
	}

	return nil
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

// listRequest what every list of events selects, and its page.
type listRequest interface {
	eventfilter.Selection
	GetPage() *typepb.Pages
}

// listEvents the events list returns for the filter of req; an unknown
// field is an invalid argument.
func listEvents[T any](ctx context.Context, req listRequest,
	list func(context.Context, *eventfilter.Filter) ([]*T, error),
) ([]*T, error) {
	filter, err := listFilter(req)
	if err != nil {
		return nil, err
	}

	events, err := list(ctx, filter)
	if err != nil {
		return nil, filterError(err)
	}

	return events, nil
}

// listFilter the filter of the list of req, its page bounded; a filter
// that is not "field:value" is an invalid argument.
func listFilter(req listRequest) (*eventfilter.Filter, error) {
	filter, err := eventfilter.FilterOf(req)
	if err != nil {
		return nil, filterError(err)
	}

	filter.Page = eventPage(req.GetPage())

	return filter, nil
}

// rangedFilter the filter of sel up to its to, else now, from its from,
// else span before: a scan never reads a whole table.
func rangedFilter(sel eventfilter.Selection, now time.Time,
	span time.Duration) (*eventfilter.Filter, error) {
	filter, err := eventfilter.FilterOf(sel)
	if err != nil {
		return nil, filterError(err)
	}

	if filter.From.IsZero() {
		end := filter.To
		if end.IsZero() {
			end = now
		}
		filter.From = end.Add(-span)
	}

	return filter, nil
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

// filterError err, an invalid argument when the request selects events
// wrongly: a field the events do not have, a filter that is not
// "field:value", or a value that is not one of the type of its field.
func filterError(err error) error {
	switch {
	case errors.Is(err, eventfilter.ErrUnknownField),
		errors.Is(err, eventfilter.ErrInvalidFilter),
		errors.Is(err, eventfilter.ErrInvalidValue):
		return szerror.StatusInvalidArgumentF("%s", err)
	default:
		return err
	}
}
