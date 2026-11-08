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
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	"github.com/sentinez/pkg/sztable"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

const _totalColumn = "total"

// _sumTotal the events of the rows of a group of a view.
const _sumTotal = "sum(" + _totalColumn + ")::BIGINT"

// _secondRetention how long the Second buckets are kept: they are many,
// and only the latest ones are charted.
const _secondRetention = 7 * 24 * time.Hour

// _secondSpan the longest range charted per Second when the Resolution
// is not set: a zoom into the minutes (Filter.resolution).
const _secondSpan = 15 * time.Minute

// Resolution the time range of the buckets of a histogram; numbered as
// analyticpb.HistogramResolution, so the api converts it as is.
type Resolution int

const (
	// Second 60 buckets of a second.
	Second Resolution = iota + 1
	// Minute 60 buckets of a minute.
	Minute
	// Hour 24 buckets of an hour.
	Hour
)

// resolutionSpec the view of a Resolution: its name suffix, bucket,
// number of buckets per query, refresh policy and retention, that of the
// source when zero. The late events of kafka are counted within
// StartOffset; the latest buckets, not materialized yet, are aggregated
// on query (RealTime).
type resolutionSpec struct {
	suffix    string
	bucket    time.Duration
	count     int
	refresh   timescale.RefreshPolicy
	retention time.Duration
}

// _resolutions the view of each Resolution.
var _resolutions = map[Resolution]resolutionSpec{
	Second: {
		suffix: "secondly",
		bucket: time.Second,
		count:  60,
		refresh: timescale.RefreshPolicy{
			StartOffset: 3 * time.Minute,
			EndOffset:   time.Second,
			Schedule:    10 * time.Second,
		},
		retention: _secondRetention,
	},
	Minute: {
		suffix: "minutely",
		bucket: time.Minute,
		count:  60,
		refresh: timescale.RefreshPolicy{
			StartOffset: 2 * time.Hour,
			EndOffset:   time.Minute,
			Schedule:    time.Minute,
		},
	},
	Hour: {
		suffix: "hourly",
		bucket: time.Hour,
		count:  24,
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

// Point the number of events of a resource and key, or value, in the
// bucket starting at Time.
type Point struct {
	Time       time.Time
	ResourceId string
	// Key the value of Spec.Key, 0 without.
	Key int32
	// Value the value of ValueSpec.Column, empty for a Spec.
	Value string
	Total int64
}

// Filter selects the points of a histogram, oldest first; a zero field
// does not filter.
type Filter struct {
	// Resolution picked by the range when zero (Filter.resolution).
	Resolution  Resolution
	ResourceIds []string
	TenantId    string
	// From and To the range picking the Resolution when zero. The buckets
	// start before To (exclusive), else end with the current one; From
	// only picks the Resolution.
	From time.Time
	To   time.Time
}

// Histogram the continuous aggregates of a Spec.
type Histogram struct {
	views views
	key   string
}

// views the continuous aggregates of a histogram, one per Resolution.
type views map[Resolution]*timescale.View[Point]

// viewDef the views of a histogram, but their resolution: the events
// counted per bucket, resource, tenant then groupBy.
type viewDef struct {
	source timescale.Hypertable
	// name inserted in the view names, <source>.<name>.minutely: the
	// views of one source with another groupBy need one.
	name    string
	groupBy []string
}

// window the buckets a query reads: count buckets of bucket, the first
// starting at from.
type window struct {
	from   time.Time
	bucket time.Duration
	count  int
}

// series the points of a resource and key.
type series struct {
	resourceId string
	key        int32
}

// New returns the Histogram of spec, its views, named after the source
// table and their resolution (<source>.minutely), synced on the way; its
// source hypertable must exist.
func New(ctx context.Context, conf *settingpb.Config,
	spec Spec) (*Histogram, error) {

	views, err := newViews(ctx, conf, spec.def())
	if err != nil {
		return nil, err
	}

	return &Histogram{views: views, key: spec.Key}, nil
}

// Points the points of filter, nil included, oldest first: one per
// bucket, resource and key, 0 when the bucket has no event, a resource of
// the filter without any counted with key 0.
func (h *Histogram) Points(ctx context.Context,
	filter *Filter) ([]*Point, error) {

	view, w, err := h.views.of(filter, time.Now())
	if err != nil {
		return nil, err
	}

	points, err := view.CollectRows(ctx,
		pointsQuery(view.Table(), h.key, filter, w), scanKey)
	if err != nil {
		return nil, err
	}

	return w.fill(points, filter.resourceIds()), nil
}

// def the views of s.
func (s Spec) def() viewDef {
	def := viewDef{source: s.Source}
	if s.Key != "" {
		def.groupBy = []string{s.Key}
	}

	return def
}

// newViews creates, or syncs, the views of def, one per Resolution.
func newViews(ctx context.Context, conf *settingpb.Config,
	def viewDef) (views, error) {

	v := make(views, len(_resolutions))
	for res := range _resolutions {
		view, err := timescale.NewView[Point](ctx, conf, def.schema(res))
		if err != nil {
			return nil, err
		}
		v[res] = view
	}

	return v, nil
}

// of the view and window of the resolution of filter, nil included.
func (v views) of(filter *Filter,
	now time.Time) (*timescale.View[Point], window, error) {

	res := filter.resolution(now)
	view, ok := v[res]
	if !ok {
		return nil, window{}, fmt.Errorf(
			"histogram: unknown resolution %d", res)
	}

	return view, filter.window(_resolutions[res], now), nil
}

// resolution the Resolution of f, nil included: the one set, else Second
// for a bounded range of up to _secondSpan that the seconds still keep,
// else Minute.
func (f *Filter) resolution(now time.Time) Resolution {
	switch {
	case f == nil:
		return Minute
	case f.Resolution != 0:
		return f.Resolution
	case f.From.IsZero():
		return Minute
	}

	to := f.To
	if to.IsZero() {
		to = now
	}

	span := to.Sub(f.From)
	if span <= 0 || span > _secondSpan ||
		now.Sub(f.From) > _secondRetention {
		return Minute
	}

	return Second
}

// window the last buckets of spec f reads, nil included: those starting
// before To when set, else up to the current one, that of now.
func (f *Filter) window(spec resolutionSpec, now time.Time) window {
	// the end of the current bucket, else To: the buckets start before it
	end := now.Truncate(spec.bucket).Add(spec.bucket)
	if f != nil && !f.To.IsZero() {
		end = ceil(f.To, spec.bucket)
	}
	span := time.Duration(spec.count) * spec.bucket

	return window{
		from:   end.Add(-span).UTC(),
		bucket: spec.bucket,
		count:  spec.count,
	}
}

// resourceIds the resources of f, nil included.
func (f *Filter) resourceIds() []string {
	if f == nil {
		return nil
	}

	return f.ResourceIds
}

// to the end of w, exclusive.
func (w window) to() time.Time {
	return w.from.Add(time.Duration(w.count) * w.bucket)
}

// fill the points of w, oldest first then per resource and key: those of
// points, a zero one for each bucket of a series without; the series
// those of points and resourceIds, the latter with key 0 when without.
func (w window) fill(points []*Point, resourceIds []string) []*Point {
	type slot struct {
		series
		millis int64
	}

	found := make(map[slot]*Point, len(points))
	for _, p := range points {
		s := series{resourceId: p.ResourceId, key: p.Key}
		found[slot{series: s, millis: p.Time.UnixMilli()}] = p
	}

	all := seriesOf(points, resourceIds)
	filled := make([]*Point, 0, w.count*len(all))
	for i := range w.count {
		start := w.from.Add(time.Duration(i) * w.bucket)
		for _, s := range all {
			p, ok := found[slot{series: s, millis: start.UnixMilli()}]
			if !ok {
				p = &Point{Time: start, ResourceId: s.resourceId, Key: s.key}
			}
			filled = append(filled, p)
		}
	}

	return filled
}

// seriesOf the series of points, then those of resourceIds without any
// point, with key 0; sorted by resource then key.
func seriesOf(points []*Point, resourceIds []string) []series {
	all := make([]series, 0, len(points)+len(resourceIds))
	counted := make(map[string]bool, len(resourceIds))
	for _, p := range points {
		counted[p.ResourceId] = true
		all = append(all, series{resourceId: p.ResourceId, key: p.Key})
	}
	for _, id := range resourceIds {
		if !counted[id] {
			all = append(all, series{resourceId: id})
		}
	}

	slices.SortFunc(all, func(a, b series) int {
		return cmp.Or(cmp.Compare(a.resourceId, b.resourceId),
			cmp.Compare(a.key, b.key))
	})

	return slices.Compact(all)
}

// ceil the first multiple of d from t on.
func ceil(t time.Time, d time.Duration) time.Time {
	truncated := t.Truncate(d)
	if truncated.Before(t) {
		return truncated.Add(d)
	}

	return truncated
}

// schema the view of d at res: the events counted per bucket, resource,
// tenant then groupBy, kept for the retention of the resolution.
func (d viewDef) schema(res Resolution) timescale.ViewSchema {
	spec := _resolutions[res]
	groupBy := append([]string{typepb.Owner_ResourceId,
		typepb.Owner_TenantId}, d.groupBy...)

	name := d.source.Schema.Table.String() + "."
	if d.name != "" {
		name += d.name + "."
	}

	return timescale.ViewSchema{
		Name:       sztable.Table(name + spec.suffix),
		Source:     d.source,
		Bucket:     spec.bucket,
		GroupBy:    groupBy,
		Aggregates: []timescale.Aggregate{{Expr: "count(*)", As: _totalColumn}},
		RealTime:   true,
		Refresh:    spec.refresh,
		// the retention of the resolution, else that of the source, so
		// the buckets live as long as the events they count
		Retention: cmp.Or(spec.retention, d.source.Retention),
	}
}

// pointsQuery sums the totals of the view per bucket, resource and key:
// over the tenants, one per resource, of the filter and w.
func pointsQuery(view, key string, filter *Filter,
	w window) sq.SelectBuilder {

	return bucketQuery(view, key, _sumTotal, filter, w)
}

// bucketQuery selects total per bucket, resource and column of the view,
// of the filter and w, oldest first; the column is 0 when empty.
func bucketQuery(view, column, total string, filter *Filter,
	w window) sq.SelectBuilder {

	selected := "0"
	groupBy := []string{timescale.BucketColumn, typepb.Owner_ResourceId}
	if column != "" {
		selected = column
		groupBy = append(groupBy, column)
	}

	b := sq.Select(timescale.BucketColumn, typepb.Owner_ResourceId, selected,
		total).
		From(view).
		Where(sq.GtOrEq{timescale.BucketColumn: w.from.UnixMilli()}).
		Where(sq.Lt{timescale.BucketColumn: w.to().UnixMilli()}).
		GroupBy(groupBy...).
		OrderBy(timescale.BucketColumn, typepb.Owner_ResourceId)

	return filter.apply(b)
}

// apply adds the resource and tenant conditions of f, nil included, to b.
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

	return b
}

// scanKey scans the rows of a bucketQuery of an INTEGER column, or of
// none, into Point.Key.
func scanKey(rows dbx.Rows) ([]*Point, error) {
	return scanPoints(rows, func(p *Point) any { return &p.Key })
}

// scanValue scans the rows of a bucketQuery of a TEXT column into
// Point.Value.
func scanValue(rows dbx.Rows) ([]*Point, error) {
	return scanPoints(rows, func(p *Point) any { return &p.Value })
}

// scanPoints scans the rows of a bucketQuery, its column into the field
// of the Point that column returns.
func scanPoints(rows dbx.Rows, column func(*Point) any) ([]*Point, error) {
	var points []*Point
	for rows.Next() {
		var (
			point  Point
			millis int64
		)

		err := rows.Scan(&millis, &point.ResourceId, column(&point),
			&point.Total)
		if err != nil {
			return nil, err
		}

		point.Time = time.UnixMilli(millis).UTC()
		points = append(points, &point)
	}

	return points, rows.Err()
}
