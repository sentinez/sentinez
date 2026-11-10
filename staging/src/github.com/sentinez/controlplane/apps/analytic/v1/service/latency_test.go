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
	"testing"
	"time"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLatencyFilter(t *testing.T) {
	now := time.UnixMilli(10_000_000_000).UTC()
	from := now.Add(-time.Hour)
	to := now.Add(-time.Minute)

	tests := []struct {
		name     string
		give     *pb.EventSelection
		wantFrom time.Time
		wantTo   time.Time
	}{
		{
			name:     "unset: the last day up to now",
			wantFrom: now.Add(-_latencyRange),
			wantTo:   now,
		},
		{
			name:     "to only: the day before it",
			give:     &pb.EventSelection{To: timestamppb.New(to)},
			wantFrom: to.Add(-_latencyRange),
			wantTo:   to,
		},
		{
			name: "range",
			give: &pb.EventSelection{
				From: timestamppb.New(from),
				To:   timestamppb.New(to),
			},
			wantFrom: from,
			wantTo:   to,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := latencyFilter(tt.give, now)

			require.NoError(t, err)
			assert.Equal(t, tt.wantFrom, got.From)
			assert.Equal(t, tt.wantTo, got.To)
		})
	}
}

func TestLatencyFilterNotKept(t *testing.T) {
	now := time.UnixMilli(10_000_000_000).UTC()
	give := &pb.EventSelection{
		From: timestamppb.New(now.Add(-60 * 24 * time.Hour)),
	}

	_, err := latencyFilter(give, now)

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestLatencyFilterRejectsFilter(t *testing.T) {
	_, err := latencyFilter(&pb.EventSelection{Filters: []string{"x"}},
		time.Now())

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestLatencyResponse(t *testing.T) {
	now := time.UnixMilli(10_000_000_000).UTC()
	filter, err := latencyFilter(nil, now)
	require.NoError(t, err)

	got := latencyResponse(filter, []*pb.UpstreamLatency{
		{
			Upstream: "http://a", Total: 3, AvgMicros: 100,
			P50Micros: 90, P95Micros: 150, P99Micros: 190, MaxMicros: 200,
		},
	})

	require.Len(t, got.GetUpstreams(), 1)
	upstream := got.GetUpstreams()[0]
	assert.Equal(t, "http://a", upstream.GetUpstream())
	assert.Equal(t, int64(3), upstream.GetTotal())
	assert.Equal(t, int64(100), upstream.GetAvgMicros())
	assert.Equal(t, int64(90), upstream.GetP50Micros())
	assert.Equal(t, int64(150), upstream.GetP95Micros())
	assert.Equal(t, int64(190), upstream.GetP99Micros())
	assert.Equal(t, int64(200), upstream.GetMaxMicros())
	assert.Equal(t, filter.From, got.GetFrom().AsTime())
	assert.Equal(t, now, got.GetTo().AsTime())
}

func TestLatencyResponseEmpty(t *testing.T) {
	got := latencyResponse(&eventfilter.Filter{}, nil)

	assert.Empty(t, got.GetUpstreams())
}
