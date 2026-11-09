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
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/pkg/szstorage/dbx/timescale/histogram"
)

// Points the points of the histogram h of filter, nil included (see
// histogram.Histogram.Points): of its views when f is nil, else of the
// events of its source f selects, on cols: f should hold the requests and
// matches only, as the resources and time range are those of filter. A
// Match of a field cols does not have is an ErrUnknownField.
func Points(ctx context.Context, h *histogram.Histogram,
	filter *histogram.Filter, f *Filter,
	cols Columns) ([]*histogram.Point, error) {
	if f == nil {
		return h.Points(ctx, filter)
	}

	return h.PointsWhere(ctx, filter,
		func(b sq.SelectBuilder) (sq.SelectBuilder, error) {
			return f.Where(b, cols)
		})
}
