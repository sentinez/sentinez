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

package eventfilter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// _selection a Selection of fixed getters.
type _selection struct {
	filters []string
	from    *timestamppb.Timestamp
}

func (s _selection) GetResourceIds() []string { return []string{"r1"} }
func (s _selection) GetRequestIds() []string  { return []string{"req_1"} }
func (s _selection) GetFilters() []string     { return s.filters }

func (s _selection) GetFrom() *timestamppb.Timestamp { return s.from }
func (s _selection) GetTo() *timestamppb.Timestamp   { return nil }

func TestFilterOf(t *testing.T) {
	give := _selection{
		filters: []string{"status:429"},
		from:    timestamppb.New(time.UnixMilli(1000)),
	}

	got, err := FilterOf(give)

	require.NoError(t, err)
	assert.Equal(t, &Filter{
		RequestIds:  []string{"req_1"},
		ResourceIds: []string{"r1"},
		Matches:     []Match{{Field: "status", Value: "429"}},
		From:        time.UnixMilli(1000).UTC(),
	}, got)
}

func TestFilterOfInvalid(t *testing.T) {
	_, err := FilterOf(_selection{filters: []string{"nocolon"}})

	assert.ErrorIs(t, err, ErrInvalidFilter)
}

func TestMatchesOf(t *testing.T) {
	got, err := MatchesOf([]string{"status:429", "path:/a:b", "q: x "})

	require.NoError(t, err)
	assert.Equal(t, []Match{
		{Field: "status", Value: "429"},
		{Field: "path", Value: "/a:b"},
		{Field: "q", Value: " x "},
	}, got)
}

func TestMatchesOfNone(t *testing.T) {
	got, err := MatchesOf(nil)

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestMatchesOfRejectsFilter(t *testing.T) {
	for _, give := range []string{"status", ":429", ""} {
		t.Run(give, func(t *testing.T) {
			_, err := MatchesOf([]string{give})

			assert.ErrorIs(t, err, ErrInvalidFilter)
		})
	}
}

func TestTimeOf(t *testing.T) {
	assert.True(t, TimeOf(nil).IsZero())
	assert.Equal(t, time.UnixMilli(5).UTC(),
		TimeOf(timestamppb.New(time.UnixMilli(5))))
}
