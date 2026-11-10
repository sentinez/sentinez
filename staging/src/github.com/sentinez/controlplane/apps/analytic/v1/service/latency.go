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
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// _latencyRange the range of a latency without from: the day before its
// end.
const _latencyRange = 24 * time.Hour

// GetEgressLatency implements pb.AnalyticServiceServer.
func (srv *AnalyticService) GetEgressLatency(ctx context.Context,
	req *pb.GetEgressLatencyRequest) (*pb.GetEgressLatencyResponse, error) {
	filter, err := latencyFilter(req.GetSelection(), time.Now())
	if err != nil {
		return nil, err
	}

	rows, err := srv.events.HttpEgress.Latency(ctx, filter)
	if err != nil {
		return nil, filterError(err)
	}

	return latencyResponse(filter, rows), nil
}

// latencyFilter the requests sel selects at now: up to to, else now, from
// from, else _latencyRange before; from or to older than the events are
// kept is an invalid argument.
func latencyFilter(sel *pb.EventSelection,
	now time.Time) (*eventfilter.Filter, error) {
	filter, err := rangedFilter(sel, now, _latencyRange)
	if err != nil {
		return nil, err
	}

	if err := keptSince(now, filter.From, filter.To); err != nil {
		return nil, err
	}
	if filter.To.IsZero() {
		filter.To = now
	}

	return filter, nil
}

// latencyResponse the response of rows, the latency over the range of
// filter.
func latencyResponse(filter *eventfilter.Filter,
	rows []*pb.UpstreamLatency) *pb.GetEgressLatencyResponse {
	return &pb.GetEgressLatencyResponse{
		Upstreams: rows,
		From:      timestamppb.New(filter.From),
		To:        timestamppb.New(filter.To),
	}
}
