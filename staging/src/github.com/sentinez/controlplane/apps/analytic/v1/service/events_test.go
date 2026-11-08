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

	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestResolutionMatchesProto(t *testing.T) {
	tests := []struct {
		give pb.HistogramResolution
		want histogram.Resolution
	}{
		{pb.HistogramResolution_HISTOGRAM_RESOLUTION_UNSPECIFIED, 0},
		{pb.HistogramResolution_HISTOGRAM_RESOLUTION_SECOND, histogram.Second},
		{pb.HistogramResolution_HISTOGRAM_RESOLUTION_MINUTE, histogram.Minute},
		{pb.HistogramResolution_HISTOGRAM_RESOLUTION_HOUR, histogram.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.give.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, histogram.Resolution(tt.give))
		})
	}
}

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
	from := time.UnixMilli(1000).UTC()

	tests := []struct {
		name string
		give histogramRequest
		want *histogram.Filter
	}{
		{
			name: "event histogram",
			give: &pb.GetEventHistogramRequest{
				Resolution:  pb.HistogramResolution_HISTOGRAM_RESOLUTION_HOUR,
				ResourceIds: []string{"r1"},
				From:        timestamppb.New(from),
			},
			want: &histogram.Filter{
				Resolution:  histogram.Hour,
				ResourceIds: []string{"r1"},
				From:        from,
			},
		},
		{
			name: "visitors",
			give: &pb.ListVisitorsRequest{},
			want: &histogram.Filter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, histogramFilter(tt.give))
		})
	}
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
