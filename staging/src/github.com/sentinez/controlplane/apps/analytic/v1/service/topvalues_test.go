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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/topvalues"
	pb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTopLimit(t *testing.T) {
	tests := []struct {
		name string
		give int32
		want uint64
	}{
		{"unset", 0, _topLimit},
		{"negative", -1, _topLimit},
		{"within", 10, 10},
		{"too large", 1000, _maxTopLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, topLimit(tt.give))
		})
	}
}

func TestTopFilter(t *testing.T) {
	now := time.UnixMilli(10_000_000).UTC()
	from := now.Add(-time.Minute)
	to := now.Add(-time.Second)

	tests := []struct {
		name string
		give *pb.GetEventTopValuesRequest
		want *eventfilter.Filter
	}{
		{
			name: "unbounded",
			give: &pb.GetEventTopValuesRequest{},
			want: &eventfilter.Filter{From: now.Add(-_topRange)},
		},
		{
			name: "to only",
			give: &pb.GetEventTopValuesRequest{To: timestamppb.New(to)},
			want: &eventfilter.Filter{From: to.Add(-_topRange), To: to},
		},
		{
			name: "range",
			give: &pb.GetEventTopValuesRequest{
				ResourceIds: []string{"r1"},
				From:        timestamppb.New(from),
				To:          timestamppb.New(to),
			},
			want: &eventfilter.Filter{
				ResourceIds: []string{"r1"},
				From:        from,
				To:          to,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := topFilter(tt.give, now)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTopValuesResponse(t *testing.T) {
	result := &topvalues.Result{
		Total: 4,
		Fields: []*topvalues.FieldValues{
			{Field: "method", Values: []*topvalues.Value{
				{Value: "GET", Total: 3},
				{Value: "POST", Total: 1},
			}},
			{Field: "country"},
		},
	}

	got := topValuesResponse(result)

	assert.Equal(t, int64(4), got.GetTotal())
	require.Len(t, got.GetFields(), 2)
	method := got.GetFields()[0]
	assert.Equal(t, "method", method.GetField())
	require.Len(t, method.GetValues(), 2)
	assert.Equal(t, "GET", method.GetValues()[0].GetValue())
	assert.Equal(t, int64(3), method.GetValues()[0].GetTotal())
	assert.InDelta(t, 0.75, method.GetValues()[0].GetShare(), 1e-9)
	assert.Equal(t, "country", got.GetFields()[1].GetField())
	assert.Empty(t, got.GetFields()[1].GetValues())
}

func TestShareOf(t *testing.T) {
	assert.InDelta(t, 0.5, shareOf(1, 2), 1e-9)
	assert.Zero(t, shareOf(1, 0))
}

func TestMatchesOf(t *testing.T) {
	tests := []struct {
		name string
		give []string
		want []eventfilter.Match
	}{
		{"none", nil, nil},
		{
			"value with colons and spaces",
			[]string{"user_agent:Mozilla/5.0 (X11; a:b)", "status:429"},
			[]eventfilter.Match{
				{Field: "user_agent", Value: "Mozilla/5.0 (X11; a:b)"},
				{Field: "status", Value: "429"},
			},
		},
		{"empty value", []string{"country:"}, []eventfilter.Match{
			{Field: "country", Value: ""},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchesOf(tt.give)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchesOfRejectsFilter(t *testing.T) {
	tests := []string{"status", ":429"}

	for _, give := range tests {
		t.Run(give, func(t *testing.T) {
			_, err := matchesOf([]string{give})

			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestFieldError(t *testing.T) {
	unknown := fmt.Errorf("where: %w", eventfilter.ErrUnknownField)
	other := errors.New("db down")

	assert.Equal(t, codes.InvalidArgument, status.Code(fieldError(unknown)))
	assert.Equal(t, other, fieldError(other))
}
