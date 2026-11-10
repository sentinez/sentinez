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
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ErrInvalidFilter the error of a field filter that is not "field:value".
var ErrInvalidFilter = errors.New("invalid filter: want field:value")

// Selection what a request selects of the events: the requests of the
// lists, the histogram and the top values, and the EventSelection message,
// all have these getters.
type Selection interface {
	GetResourceIds() []string
	GetRequestIds() []string
	GetFilters() []string
	GetFrom() *timestamppb.Timestamp
	GetTo() *timestamppb.Timestamp
}

// FilterOf the Filter of sel, without page; an ErrInvalidFilter when one
// of its filters is not "field:value".
func FilterOf(sel Selection) (*Filter, error) {
	matches, err := MatchesOf(sel.GetFilters())
	if err != nil {
		return nil, err
	}

	return &Filter{
		RequestIds:  sel.GetRequestIds(),
		ResourceIds: sel.GetResourceIds(),
		Matches:     matches,
		From:        TimeOf(sel.GetFrom()),
		To:          TimeOf(sel.GetTo()),
	}, nil
}

// MatchesOf the Match of each of filters, each "field:value": the field
// is before the first colon, the value is the rest. A filter without a
// field is an ErrInvalidFilter.
func MatchesOf(filters []string) ([]Match, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	matches := make([]Match, 0, len(filters))
	for _, f := range filters {
		field, value, ok := strings.Cut(f, ":")
		if !ok || field == "" {
			return nil, fmt.Errorf("%w: %q", ErrInvalidFilter, f)
		}
		matches = append(matches, Match{Field: field, Value: value})
	}

	return matches, nil
}

// TimeOf the time of ts, zero when unset.
func TimeOf(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}

	return ts.AsTime()
}
