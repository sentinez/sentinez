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
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEventPage(t *testing.T) {
	tests := []struct {
		name string
		give *typepb.Pages
		want *typepb.Pages
	}{
		{"unset", nil, &typepb.Pages{Size: _pageSize, Index: 1}},
		{
			"no index",
			&typepb.Pages{Size: 10},
			&typepb.Pages{Size: 10, Index: 1},
		},
		{
			"too large",
			&typepb.Pages{Size: 10_000, Index: 3},
			&typepb.Pages{Size: _maxPageSize, Index: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eventPage(tt.give)

			assert.Equal(t, tt.want.GetSize(), got.GetSize())
			assert.Equal(t, tt.want.GetIndex(), got.GetIndex())
		})
	}
}

func TestHistogramFilter(t *testing.T) {
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	from := now.Add(-time.Hour)

	tests := []struct {
		name         string
		give         histogramRequest
		want         *histogram.Filter
		wantInterval typepb.HistogramInterval
	}{
		{
			name: "event histogram",
			give: &pb.GetEventHistogramRequest{
				Resolution:  typepb.HistogramResolution_HISTOGRAM_RESOLUTION_1_HOUR,
				ResourceIds: []string{"r1"},
				From:        timestamppb.New(from),
			},
			want: &histogram.Filter{
				Resolution:  typepb.HistogramResolution_HISTOGRAM_RESOLUTION_1_HOUR,
				ResourceIds: []string{"r1"},
				From:        from,
			},
			wantInterval: typepb.HistogramInterval_HISTOGRAM_INTERVAL_1_MINUTE,
		},
		{
			name: "visitors, resolution picked",
			give: &pb.ListVisitorsRequest{},
			want: &histogram.Filter{
				Resolution: typepb.HistogramResolution_HISTOGRAM_RESOLUTION_1_HOUR,
			},
			wantInterval: typepb.HistogramInterval_HISTOGRAM_INTERVAL_1_MINUTE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, interval, err := histogramFilter(tt.give, now)

			require.NoError(t, err)
			assert.Equal(t, tt.want, filter)
			assert.Equal(t, tt.wantInterval, interval)
		})
	}
}

func TestHistogramFilterNotKept(t *testing.T) {
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	old := timestamppb.New(now.Add(-31 * 24 * time.Hour))

	tests := []struct {
		name string
		give histogramRequest
	}{
		{"from", &pb.GetEventHistogramRequest{From: old}},
		{"to", &pb.ListVisitorsRequest{To: old}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := histogramFilter(tt.give, now)

			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestHistogramFilterKept(t *testing.T) {
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	req := &pb.GetEventHistogramRequest{
		From: timestamppb.New(now.Add(-30 * 24 * time.Hour)),
	}

	_, interval, err := histogramFilter(req, now)

	require.NoError(t, err)
	assert.Equal(t,
		typepb.HistogramInterval_HISTOGRAM_INTERVAL_12_HOURS, interval)
}

func TestHistogramPoints(t *testing.T) {
	at := time.UnixMilli(1000).UTC()

	got := histogramPoints([]*histogram.Point{
		{Time: at, ResourceId: "r1", Key: 200, Total: 3},
	})

	assert.Len(t, got, 1)
	assert.Equal(t, at, got[0].GetTime().AsTime())
	assert.Equal(t, "r1", got[0].GetResourceId())
	assert.Equal(t, int32(200), got[0].GetKey())
	assert.Equal(t, int64(3), got[0].GetTotal())
}

func TestHistogramEvents(t *testing.T) {
	tests := []struct {
		name string
		give *pb.GetEventHistogramRequest
		want *eventfilter.Filter
	}{
		{
			name: "none counts every event",
			give: &pb.GetEventHistogramRequest{},
		},
		{
			name: "requests and filters",
			give: &pb.GetEventHistogramRequest{
				RequestIds: []string{"req_1"},
				Filters:    []string{"status:404"},
			},
			want: &eventfilter.Filter{
				RequestIds: []string{"req_1"},
				Matches: []eventfilter.Match{
					{Field: "status", Value: "404"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := histogramEvents(tt.give)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHistogramEventsRejectsFilter(t *testing.T) {
	_, err := histogramEvents(&pb.GetEventHistogramRequest{
		Filters: []string{"status"},
	})

	require.Error(t, err)
}
