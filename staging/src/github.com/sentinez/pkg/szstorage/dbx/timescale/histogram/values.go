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

package histogram

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// ValueSpec describes the histograms of the values of a column of an
// event hypertable, ex: the client ips: the events counted per bucket,
// resource and value, so both the values and their number are exact.
type ValueSpec struct {
	Source timescale.Hypertable
	// Column a TEXT column of Source, ex: the client ip; it names the
	// views, <source>.<column>.minutely.
	Column string
}

// Values the continuous aggregates of a ValueSpec.
type Values struct {
	views  views
	column string
}

// NewValues returns the Values of spec, its views synced on the way; its
// source hypertable must exist.
func NewValues(ctx context.Context, conf *settingpb.Config,
	spec ValueSpec) (*Values, error) {

	views, err := newViews(ctx, conf, viewDef{
		source:  spec.Source,
		name:    spec.Column,
		groupBy: []string{spec.Column},
	})
	if err != nil {
		return nil, err
	}

	return &Values{views: views, column: spec.Column}, nil
}

// Points the events per bucket, resource and value (Point.Value) of
// filter, nil included: oldest bucket first, then the most frequent value;
// the buckets without events have none.
func (v *Values) Points(ctx context.Context,
	filter *Filter) ([]*Point, error) {

	view, w, err := v.views.of(filter, time.Now())
	if err != nil {
		return nil, err
	}

	return view.CollectRows(ctx,
		valuesQuery(view.Table(), v.column, filter, w), scanValue)
}

// Distinct the number of distinct values per bucket and resource of
// filter, nil included, oldest first: one per bucket and resource, 0 when
// the bucket has no event; Point.Key is 0.
func (v *Values) Distinct(ctx context.Context,
	filter *Filter) ([]*Point, error) {

	view, w, err := v.views.of(filter, time.Now())
	if err != nil {
		return nil, err
	}

	points, err := view.CollectRows(ctx,
		distinctQuery(view.Table(), v.column, filter, w), scanKey)
	if err != nil {
		return nil, err
	}

	return w.fill(points, filter.resourceIds()), nil
}

// valuesQuery sums the totals of the view per bucket, resource and value
// of column, the most frequent value first.
func valuesQuery(view, column string, filter *Filter,
	w window) sq.SelectBuilder {

	return pointsQuery(view, column, filter, w).
		OrderBy(_sumTotal+" DESC", column)
}

// distinctQuery counts the distinct values of column per bucket and
// resource of the view: over the tenants, a value of several counted once.
func distinctQuery(view, column string, filter *Filter,
	w window) sq.SelectBuilder {

	return bucketQuery(view, "", "count(DISTINCT "+column+")::BIGINT",
		filter, w)
}
