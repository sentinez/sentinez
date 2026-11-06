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

// Package histogram counts the events of an event hypertable per time
// bucket, resource and key, in a continuous aggregate per Resolution:
// the series of the charts of the events.
package histogram

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/pkg/schema"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/szstorage/dbx"
	"github.com/sentinez/shared/szstorage/dbx/timescale"
	"github.com/sentinez/shared/sztable"
)

const _totalColumn = "total"

// Resolution the time range of the buckets of a histogram.
type Resolution int

const (
	// Minute buckets of a minute, for ranges up to a day.
	Minute Resolution = iota + 1
	// Hour buckets of an hour, for ranges up to EventRetention.
	Hour
)

// _resolutions the view of each Resolution, in order: its name suffix,
// bucket and refresh policy. The late events of kafka are counted within
// StartOffset; the latest buckets, not materialized yet, are aggregated
// on query (RealTime).
var _resolutions = []struct {
	resolution Resolution
	suffix     string
	bucket     time.Duration
	refresh    timescale.RefreshPolicy
}{
	{
		resolution: Minute,
		suffix:     "minutely",
		bucket:     time.Minute,
		refresh: timescale.RefreshPolicy{
			StartOffset: 2 * time.Hour,
			EndOffset:   time.Minute,
			Schedule:    time.Minute,
		},
	},
	{
		resolution: Hour,
		suffix:     "hourly",
		bucket:     time.Hour,
		refresh: timescale.RefreshPolicy{
			StartOffset: 3 * 24 * time.Hour,
			EndOffset:   time.Hour,
			Schedule:    30 * time.Minute,
		},
	},
}

// Spec describes the histograms of an event hypertable.
type Spec struct {
	Source timescale.Hypertable
	// Key an INTEGER column of Source the points are split by, besides
	// the resource, ex: the http status; none when empty.
	Key string
}

// Point the number of events of a resource and key in the bucket
// starting at Time.
type Point struct {
	Time       time.Time
	ResourceId string
	// Key the value of Spec.Key, 0 without.
	Key   int32
	Total int64
}

// Filter selects the points of a histogram, oldest first; a zero field
// does not filter.
type Filter struct {
	// Resolution Minute when zero.
	Resolution  Resolution
	ResourceIds []string
	TenantId    string
	// From and To bound the start of the buckets, From inclusive, To
	// exclusive.
	From time.Time
	To   time.Time
}

// Histogram the continuous aggregates of a Spec.
type Histogram struct {
	views map[Resolution]*timescale.View[Point]
	key   string
}

// New returns the Histogram of spec, its views, named after the source
// table and their resolution (<source>.minutely), synced on the way; its
// source hypertable must exist.
func New(ctx context.Context, conf *settingpb.Config,
	spec Spec) (*Histogram, error) {

	views := make(map[Resolution]*timescale.View[Point], len(_resolutions))
	for i := range _resolutions {
		view, err := timescale.NewView[Point](ctx, conf, viewSchema(spec, i))
		if err != nil {
			return nil, err
		}
		views[_resolutions[i].resolution] = view
	}

	return &Histogram{views: views, key: spec.Key}, nil
}

// Points the points of filter, nil included, oldest first.
func (h *Histogram) Points(ctx context.Context,
	filter *Filter) ([]*Point, error) {

	resolution := Minute
	if filter != nil && filter.Resolution != 0 {
		resolution = filter.Resolution
	}

	view, ok := h.views[resolution]
	if !ok {
		return nil, fmt.Errorf("histogram: unknown resolution %d",
			resolution)
	}

	return view.CollectRows(ctx, pointsQuery(view.Table(), h.key, filter),
		scan)
}

// viewSchema the view of spec at _resolutions[i]: the events counted per
// bucket, resource, tenant and key, kept for EventRetention.
func viewSchema(spec Spec, i int) timescale.ViewSchema {
	res := _resolutions[i]
	groupBy := []string{typepb.Owner_ResourceId, typepb.Owner_TenantId}
	if spec.Key != "" {
		groupBy = append(groupBy, spec.Key)
	}

	return timescale.ViewSchema{
		Name: sztable.Table(spec.Source.Schema.Table.String() + "." +
			res.suffix),
		Source:     spec.Source,
		Bucket:     res.bucket,
		GroupBy:    groupBy,
		Aggregates: []timescale.Aggregate{{Expr: "count(*)", As: _totalColumn}},
		RealTime:   true,
		Refresh:    res.refresh,
		Retention:  schema.EventRetention,
	}
}

// pointsQuery sums the totals of the view per bucket, resource and key:
// over the tenants, one per resource, the filter aside.
func pointsQuery(view, key string, filter *Filter) sq.SelectBuilder {
	keyColumn := "0"
	groupBy := []string{timescale.BucketColumn, typepb.Owner_ResourceId}
	if key != "" {
		keyColumn = key
		groupBy = append(groupBy, key)
	}

	b := sq.Select(timescale.BucketColumn, typepb.Owner_ResourceId, keyColumn,
		"sum("+_totalColumn+")::BIGINT").
		From(view).
		GroupBy(groupBy...).
		OrderBy(timescale.BucketColumn, typepb.Owner_ResourceId)

	return filter.apply(b)
}

// apply adds the conditions of f, nil included, to b.
func (f *Filter) apply(b sq.SelectBuilder) sq.SelectBuilder {
	if f == nil {
		return b
	}

	if len(f.ResourceIds) > 0 {
		b = b.Where(sq.Eq{typepb.Owner_ResourceId: f.ResourceIds})
	}
	if f.TenantId != "" {
		b = b.Where(sq.Eq{typepb.Owner_TenantId: f.TenantId})
	}
	if !f.From.IsZero() {
		b = b.Where(sq.GtOrEq{timescale.BucketColumn: f.From.UnixMilli()})
	}
	if !f.To.IsZero() {
		b = b.Where(sq.Lt{timescale.BucketColumn: f.To.UnixMilli()})
	}

	return b
}

func scan(rows dbx.Rows) ([]*Point, error) {
	var points []*Point
	for rows.Next() {
		var (
			point  Point
			millis int64
		)

		err := rows.Scan(&millis, &point.ResourceId, &point.Key,
			&point.Total)
		if err != nil {
			return nil, err
		}

		point.Time = time.UnixMilli(millis).UTC()
		points = append(points, &point)
	}

	return points, rows.Err()
}
