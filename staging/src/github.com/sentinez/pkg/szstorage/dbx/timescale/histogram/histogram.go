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
// bucket, resource and key, in a continuous aggregate per interval:
// the series of the charts of the events.
package histogram

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	"github.com/sentinez/pkg/sztable"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

const _totalColumn = "total"

// errTimeColumn the source of a Histogram counted by PointsWhere has no
// time column.
var errTimeColumn = errors.New("histogram: source without time column")

// _sumTotal the events of the rows of a group of a view.
const _sumTotal = "sum(" + _totalColumn + ")::BIGINT"

// _secondRetention how long the buckets of seconds are kept: they are
// many, and only the latest ones are charted.
const _secondRetention = 7 * 24 * time.Hour

// _lateEvents how late the events of Kafka are still counted by the views
// of seconds and minutes, e.g. the backlog of a collector restart.
const _lateEvents = 2 * time.Hour

// The Resolutions and intervals, named by their span: _res<span> and
// _int<span>, s m h d the unit.
const (
	_res5m  = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_5_MINUTES
	_res15m = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_15_MINUTES
	_res30m = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_30_MINUTES
	_res1h  = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_1_HOUR
	_res3h  = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_3_HOURS
	_res6h  = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_6_HOURS
	_res12h = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_12_HOURS
	_res24h = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_24_HOURS
	_res2d  = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_2_DAYS
	_res7d  = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_7_DAYS
	_res30d = typepb.HistogramResolution_HISTOGRAM_RESOLUTION_30_DAYS

	_int10s = typepb.HistogramInterval_HISTOGRAM_INTERVAL_10_SECONDS
	_int30s = typepb.HistogramInterval_HISTOGRAM_INTERVAL_30_SECONDS
	_int1m  = typepb.HistogramInterval_HISTOGRAM_INTERVAL_1_MINUTE
	_int5m  = typepb.HistogramInterval_HISTOGRAM_INTERVAL_5_MINUTES
	_int10m = typepb.HistogramInterval_HISTOGRAM_INTERVAL_10_MINUTES
	_int15m = typepb.HistogramInterval_HISTOGRAM_INTERVAL_15_MINUTES
	_int30m = typepb.HistogramInterval_HISTOGRAM_INTERVAL_30_MINUTES
	_int1h  = typepb.HistogramInterval_HISTOGRAM_INTERVAL_1_HOUR
	_int3h  = typepb.HistogramInterval_HISTOGRAM_INTERVAL_3_HOURS
	_int12h = typepb.HistogramInterval_HISTOGRAM_INTERVAL_12_HOURS
)

// viewSpec the continuous aggregate of the buckets of a HistogramInterval:
// its name suffix, bucket, how late its events are still counted and how
// often, and its retention, that of the source when zero.
type viewSpec struct {
	suffix    string
	bucket    time.Duration
	late      time.Duration
	schedule  time.Duration
	retention time.Duration
}

// resolutionSpec the range a Resolution charts and the interval of its
// buckets: as many of them as fill it.
type resolutionSpec struct {
	span     time.Duration
	interval typepb.HistogramInterval
}

// newViewSpec the view of buckets of bucket, refreshed as often as they
// fill, 10 seconds to 30 minutes; the latest ones, not materialized yet,
// are aggregated on query (RealTime). The events of a Kafka backlog are
// counted up to _lateEvents late, 3 days for the buckets of hours; not
// more: TimescaleDB invalidates the whole time range of a transaction, so
// a batch holding one old event would make every refresh recompute it all.
// Only the seconds are kept shorter, they are many.
func newViewSpec(suffix string, bucket time.Duration) viewSpec {
	spec := viewSpec{
		suffix:   suffix,
		bucket:   bucket,
		late:     _lateEvents,
		schedule: min(max(bucket, 10*time.Second), 30*time.Minute),
	}

	switch {
	case bucket < time.Minute:
		spec.retention = _secondRetention
	case bucket >= time.Hour:
		spec.late = 3 * 24 * time.Hour
	}

	return spec
}

// refresh the refresh policy of s: the buckets completed up to a minute
// ago (or a bucket, when shorter) are materialized, so the queries
// aggregate few raw events.
func (s viewSpec) refresh() timescale.RefreshPolicy {
	return timescale.RefreshPolicy{
		StartOffset: s.late,
		EndOffset:   min(s.bucket, time.Minute),
		Schedule:    s.schedule,
	}
}

// _views the view of each HistogramInterval, named by its bucket: the
// Resolutions of one interval share it.
var _views = map[typepb.HistogramInterval]viewSpec{
	_int10s: newViewSpec("10s", 10*time.Second),
	_int30s: newViewSpec("30s", 30*time.Second),
	_int1m:  newViewSpec("1m", time.Minute),
	_int5m:  newViewSpec("5m", 5*time.Minute),
	_int10m: newViewSpec("10m", 10*time.Minute),
	_int15m: newViewSpec("15m", 15*time.Minute),
	_int30m: newViewSpec("30m", 30*time.Minute),
	_int1h:  newViewSpec("1h", time.Hour),
	_int3h:  newViewSpec("3h", 3*time.Hour),
	_int12h: newViewSpec("12h", 12*time.Hour),
}

// _resolutions the range and default interval of each Resolution.
var _resolutions = map[typepb.HistogramResolution]resolutionSpec{
	_res5m:  {span: 5 * time.Minute, interval: _int10s},
	_res15m: {span: 15 * time.Minute, interval: _int30s},
	_res30m: {span: 30 * time.Minute, interval: _int1m},
	_res1h:  {span: time.Hour, interval: _int1m},
	_res3h:  {span: 3 * time.Hour, interval: _int5m},
	_res6h:  {span: 6 * time.Hour, interval: _int10m},
	_res12h: {span: 12 * time.Hour, interval: _int15m},
	_res24h: {span: 24 * time.Hour, interval: _int30m},
	_res2d:  {span: 2 * 24 * time.Hour, interval: _int1h},
	_res7d:  {span: 7 * 24 * time.Hour, interval: _int3h},
	_res30d: {span: 30 * 24 * time.Hour, interval: _int12h},
}

// _defaultResolution the Resolution of a filter without range.
const _defaultResolution = _res1h

// _ascending the Resolutions, the shortest span first.
var _ascending = slices.SortedFunc(maps.Keys(_resolutions),
	func(a, b typepb.HistogramResolution) int {
		return cmp.Compare(_resolutions[a].span, _resolutions[b].span)
	})

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
	Resolution  typepb.HistogramResolution
	ResourceIds []string
	TenantId    string
	// From and To the range picking the Resolution when zero. The buckets
	// start before To (exclusive), else end with the current one; From
	// only picks the Resolution.
	From time.Time
	To   time.Time
}

// Histogram the continuous aggregates of a Spec, and its source.
type Histogram struct {
	views views
	key   string
	// source the events PointsWhere counts, by timeColumn, unix millis.
	source     dbx.Database[Point]
	timeColumn string
}

// Where narrows the events PointsWhere counts: the conditions it adds to
// b, on the columns of the source hypertable.
type Where func(b sq.SelectBuilder) (sq.SelectBuilder, error)

// views the continuous aggregates of a histogram, one per interval.
type views map[typepb.HistogramInterval]*timescale.View[Point]

// viewDef the views of a histogram, but their resolution: the events
// counted per bucket, resource, tenant then groupBy.
type viewDef struct {
	source timescale.Hypertable
	// name inserted in the view names, <source>.<name>.1m: the
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
// table and their interval (<source>.1m), synced on the way; its
// source hypertable must exist.
func New(ctx context.Context, conf *settingpb.Config,
	spec Spec) (*Histogram, error) {

	views, err := newViews(ctx, conf, spec.def())
	if err != nil {
		return nil, err
	}

	source, err := postgres.Open[Point](ctx, conf,
		spec.Source.Schema, postgres.NoSync)
	if err != nil {
		return nil, err
	}

	return &Histogram{
		views:      views,
		key:        spec.Key,
		source:     source,
		timeColumn: spec.Source.TimeColumn,
	}, nil
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

// PointsWhere the points of filter, nil included, as Points, but of the
// events of the source hypertable where narrows: they are counted on the
// fly, not read from a view. The time column of the source must hold unix
// milliseconds.
func (h *Histogram) PointsWhere(ctx context.Context, filter *Filter,
	where Where) ([]*Point, error) {

	if h.timeColumn == "" {
		return nil, errTimeColumn
	}

	_, w, err := planOf(filter, time.Now())
	if err != nil {
		return nil, err
	}

	b, err := where(sourceQuery(h.source.Table(), h.timeColumn, h.key,
		filter, w))
	if err != nil {
		return nil, err
	}

	points, err := h.source.CollectRows(ctx, b, scanKey)
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

// newViews creates, or syncs, the views of def, one per interval.
func newViews(ctx context.Context, conf *settingpb.Config,
	def viewDef) (views, error) {

	v := make(views, len(_views))
	for interval := range _views {
		view, err := timescale.NewView[Point](ctx, conf,
			def.schema(interval))
		if err != nil {
			return nil, err
		}
		v[interval] = view
	}

	return v, nil
}

// of the view and window of the resolution of filter, nil included.
func (v views) of(filter *Filter,
	now time.Time) (*timescale.View[Point], window, error) {

	interval, w, err := planOf(filter, now)
	if err != nil {
		return nil, window{}, err
	}

	view, ok := v[interval]
	if !ok {
		return nil, window{}, fmt.Errorf(
			"histogram: no view of interval %d", interval)
	}

	return view, w, nil
}

// planOf the interval and window of the Resolution of filter, nil
// included, at now; an unknown Resolution is an error.
func planOf(filter *Filter,
	now time.Time) (typepb.HistogramInterval, window, error) {
	res := filter.resolution(now)
	spec, ok := _resolutions[res]
	if !ok {
		return 0, window{}, fmt.Errorf(
			"histogram: unknown resolution %d", res)
	}

	return spec.interval, filter.window(spec, now), nil
}

// Resolve sets f.Resolution to the one the Points of f use at now, and
// returns the interval of its buckets; an unknown Resolution has none.
func (f *Filter) Resolve(now time.Time) typepb.HistogramInterval {
	f.Resolution = f.resolution(now)

	return _resolutions[f.Resolution].interval
}

// resolution the Resolution of f, nil included: the one set, else the
// shortest that fits the range From to To (resolutionSpec.fits), else the
// longest; _defaultResolution without a range.
func (f *Filter) resolution(now time.Time) typepb.HistogramResolution {
	switch {
	case f == nil:
		return _defaultResolution
	case f.Resolution != 0:
		return f.Resolution
	case f.From.IsZero():
		return _defaultResolution
	}

	to := f.To
	if to.IsZero() {
		to = now
	}

	span := to.Sub(f.From)
	if span <= 0 {
		return _defaultResolution
	}

	age := now.Sub(f.From)
	for _, res := range _ascending {
		if _resolutions[res].fits(span, age) {
			return res
		}
	}

	return _ascending[len(_ascending)-1]
}

// fits whether s charts a range of span starting age ago: its view still
// keeps that start, and its span holds the range but one bucket, so a
// range a little longer, e.g. from the midnight 7 days ago, keeps its
// Resolution instead of the next, much longer one.
func (s resolutionSpec) fits(span, age time.Duration) bool {
	view := _views[s.interval]
	if view.retention != 0 && age > view.retention {
		return false
	}

	return s.span+view.bucket >= span
}

// window the last buckets of spec f reads, nil included: those starting
// before To when set, else up to the current one, that of now.
func (f *Filter) window(spec resolutionSpec, now time.Time) window {
	bucket := _views[spec.interval].bucket
	// the end of the current bucket, else To: the buckets start before it
	end := now.Truncate(bucket).Add(bucket)
	if f != nil && !f.To.IsZero() {
		end = ceil(f.To, bucket)
	}

	return window{
		from:   end.Add(-spec.span).UTC(),
		bucket: bucket,
		count:  int(spec.span / bucket),
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

// schema the view of d at interval: the events counted per bucket,
// resource, tenant then groupBy, kept for the retention of the view, else
// that of the source, so the buckets live as long as the events they count.
func (d viewDef) schema(interval typepb.HistogramInterval,
) timescale.ViewSchema {
	spec := _views[interval]
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
		Refresh:    spec.refresh(),
		Retention:  cmp.Or(spec.retention, d.source.Retention),
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

// sourceQuery counts the events of table per bucket of its timeColumn
// (unix millis), resource and key, of the filter and w, oldest first; the
// key is 0 when empty.
func sourceQuery(table, timeColumn, key string, filter *Filter,
	w window) sq.SelectBuilder {

	millis := w.bucket.Milliseconds()
	bucket := fmt.Sprintf("(%[1]s / %[2]d) * %[2]d", timeColumn, millis)

	b := sq.Select(bucket, typepb.Owner_ResourceId, cmp.Or(key, "0"),
		"count(*)").
		From(table).
		Where(sq.GtOrEq{timeColumn: w.from.UnixMilli()}).
		Where(sq.Lt{timeColumn: w.to().UnixMilli()}).
		GroupBy("1", "2", "3").
		OrderBy("1", "2")

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
