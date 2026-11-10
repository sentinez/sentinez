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
	"slices"
	"testing"
	"time"

	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	"github.com/sentinez/pkg/sztable"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// _window a window of the queries of the tests: [1000, 3000) millis.
var _window = window{from: time.UnixMilli(1000), bucket: time.Second,
	count: 2}

// queryTest a case of a test of the sql of a query of a view "v" over
// _window.
type queryTest struct {
	name     string
	giveKey  string
	give     *Filter
	wantSql  string
	wantArgs []any
}

func TestPointsQuery(t *testing.T) {
	for _, tt := range pointsQueryTests() {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := pointsQuery("v", tt.giveKey, tt.give,
				_window).ToSql()

			require.NoError(t, err)
			assert.Equal(t, tt.wantSql, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func pointsQueryTests() []queryTest {
	return []queryTest{
		{
			name: "no filter, no key",
			wantSql: "SELECT bucket, resource_id, 0, sum(total)::BIGINT " +
				"FROM v WHERE bucket >= ? AND bucket < ? " +
				"GROUP BY bucket, resource_id " +
				"ORDER BY bucket, resource_id",
			wantArgs: []any{int64(1000), int64(3000)},
		},
		{
			name:    "filter and key",
			giveKey: "status",
			give: &Filter{
				ResourceIds: []string{"r1", "r2"},
				TenantId:    "t1",
			},
			wantSql: "SELECT bucket, resource_id, status, " +
				"sum(total)::BIGINT FROM v " +
				"WHERE bucket >= ? AND bucket < ? " +
				"AND resource_id IN (?,?) AND tenant_id = ? " +
				"GROUP BY bucket, resource_id, status " +
				"ORDER BY bucket, resource_id",
			wantArgs: []any{int64(1000), int64(3000), "r1", "r2", "t1"},
		},
	}
}

func TestBucketQuerySource(t *testing.T) {
	src := bucketSource{
		table:  "t",
		bucket: "time_bucket(1000::BIGINT, ts)",
		time:   "ts",
	}

	for _, tt := range sourceQueryTests() {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := bucketQuery(src, tt.giveKey, "count(*)",
				tt.give, _window).ToSql()

			require.NoError(t, err)
			assert.Equal(t, tt.wantSql, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

// sourceQueryTests the cases of TestBucketQuerySource, of a table "t"
// timed by "ts".
func sourceQueryTests() []queryTest {
	return []queryTest{
		{
			name: "no filter, no key",
			wantSql: "SELECT time_bucket(1000::BIGINT, ts), " +
				"resource_id, 0, count(*) FROM t " +
				"WHERE ts >= ? AND ts < ? " +
				"GROUP BY time_bucket(1000::BIGINT, ts), resource_id " +
				"ORDER BY time_bucket(1000::BIGINT, ts), resource_id",
			wantArgs: []any{int64(1000), int64(3000)},
		},
		{
			name:    "filter and key",
			giveKey: "status",
			give:    &Filter{ResourceIds: []string{"r1"}},
			wantSql: "SELECT time_bucket(1000::BIGINT, ts), " +
				"resource_id, status, count(*) FROM t " +
				"WHERE ts >= ? AND ts < ? AND resource_id IN (?) " +
				"GROUP BY time_bucket(1000::BIGINT, ts), resource_id, " +
				"status ORDER BY time_bucket(1000::BIGINT, ts), " +
				"resource_id",
			wantArgs: []any{int64(1000), int64(3000), "r1"},
		},
	}
}

// viewSchemaTest a case of TestViewSchema.
type viewSchemaTest struct {
	give       typepb.HistogramInterval
	wantName   sztable.Table
	wantBucket time.Duration
	wantKeep   time.Duration
	wantLate   time.Duration
}

func TestViewSchema(t *testing.T) {
	spec := Spec{
		Source: timescale.Hypertable{
			Schema:     dbx.Schema{Table: "a.events"},
			TimeColumn: "ts",
			Retention:  30 * 24 * time.Hour,
		},
		Key: "status",
	}

	for _, tt := range viewSchemaTests(spec.Source.Retention) {
		t.Run(tt.wantName.String(), func(t *testing.T) {
			view := spec.def().schema(tt.give)

			assert.Equal(t, tt.wantName, view.Name)
			assert.Equal(t, tt.wantBucket, view.Bucket)
			assert.Equal(t,
				[]string{"resource_id", "tenant_id", "status"},
				view.GroupBy)
			assert.True(t, view.RealTime)
			assert.Equal(t, tt.wantKeep, view.Retention)
			assert.Equal(t, tt.wantLate, view.Refresh.StartOffset)
			assert.Equal(t, min(tt.wantBucket, time.Minute),
				view.Refresh.EndOffset)
		})
	}
}

// viewSchemaTests the views of a source kept for sourceKeep.
func viewSchemaTests(sourceKeep time.Duration) []viewSchemaTest {
	return []viewSchemaTest{
		{
			give: _int1m, wantName: "a.events.1m",
			wantBucket: time.Minute, wantKeep: sourceKeep,
			wantLate: _lateEvents,
		},
		{
			give: _int1h, wantName: "a.events.1h",
			wantBucket: time.Hour, wantKeep: sourceKeep,
			wantLate: 3 * 24 * time.Hour,
		},
		{
			give: _int10s, wantName: "a.events.10s",
			wantBucket: 10 * time.Second, wantKeep: _secondRetention,
			wantLate: _lateEvents,
		},
	}
}

// viewNameTest a case of TestViewSchemaName.
type viewNameTest struct {
	name        string
	give        viewDef
	wantName    sztable.Table
	wantGroupBy []string
}

func TestViewSchemaName(t *testing.T) {
	for _, tt := range viewNameTests() {
		t.Run(tt.name, func(t *testing.T) {
			view := tt.give.schema(_int1m)

			assert.Equal(t, tt.wantName, view.Name)
			assert.Equal(t, tt.wantGroupBy, view.GroupBy)
		})
	}
}

func viewNameTests() []viewNameTest {
	source := timescale.Hypertable{Schema: dbx.Schema{Table: "a.events"}}

	return []viewNameTest{
		{
			name:        "spec without key",
			give:        Spec{Source: source}.def(),
			wantName:    "a.events.1m",
			wantGroupBy: []string{"resource_id", "tenant_id"},
		},
		{
			name:        "spec with key",
			give:        Spec{Source: source, Key: "k"}.def(),
			wantName:    "a.events.1m",
			wantGroupBy: []string{"resource_id", "tenant_id", "k"},
		},
		{
			name: "named",
			give: viewDef{
				source: source, name: "client_ip",
				groupBy: []string{"client_ip"},
			},
			wantName: "a.events.client_ip.1m",
			wantGroupBy: []string{
				"resource_id", "tenant_id", "client_ip",
			},
		},
	}
}

func TestViewsOf(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 30, 15, 0, time.UTC)
	hour := &timescale.View[Point]{}

	view, w, err := views{_int30m: hour}.of(
		&Filter{Resolution: _res24h}, now)

	require.NoError(t, err)
	assert.Same(t, hour, view)
	assert.Equal(t, window{
		from:   time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC),
		bucket: 30 * time.Minute,
		count:  48,
	}, w)
}

func TestViewsOfUnknown(t *testing.T) {
	_, _, err := views{}.of(
		&Filter{Resolution: _res24h}, time.Now())

	require.EqualError(t, err, "histogram: no view of interval 15")
}

// windowTest a case of TestFilterWindow.
type windowTest struct {
	name     string
	give     *Filter
	giveSpec typepb.HistogramResolution
	wantFrom time.Time
	wantTo   time.Time
	// wantCount the buckets of the window
	wantCount     int
	wantTruncated bool
}

func TestFilterWindow(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 30, 15, 0, time.UTC)

	for _, tt := range slices.Concat(windowTests(), windowFromTests(),
		windowEdgeTests()) {
		t.Run(tt.name, func(t *testing.T) {
			spec := _resolutions[tt.giveSpec]
			w := tt.give.window(spec, now)

			assert.Equal(t, tt.wantFrom, w.from)
			assert.Equal(t, tt.wantTo, w.to())
			assert.Equal(t, tt.wantCount, w.count)
			assert.Equal(t, tt.wantTruncated, w.truncated)
		})
	}
}

// windowTests the windows ending with the current bucket of 12:30:15,
// before To, from From.
func windowTests() []windowTest {
	at := func(h, m, s int) time.Time {
		return time.Date(2026, 1, 2, h, m, s, 0, time.UTC)
	}

	return []windowTest{
		{
			name: "2 days until now", give: nil, giveSpec: _res2d,
			wantFrom:  at(13, 0, 0).Add(-48 * time.Hour),
			wantTo:    at(13, 0, 0),
			wantCount: 48,
		},
		{
			name: "1 hour until now", give: &Filter{}, giveSpec: _res1h,
			wantFrom: at(11, 31, 0), wantTo: at(12, 31, 0),
			wantCount: 60,
		},
		{
			name: "5 minutes until now", give: &Filter{}, giveSpec: _res5m,
			wantFrom: at(12, 25, 20), wantTo: at(12, 30, 20),
			wantCount: 30,
		},
		{
			name:     "on a bucket start, to excluded",
			give:     &Filter{To: at(10, 0, 0)},
			giveSpec: _res1h,
			wantFrom: at(9, 0, 0), wantTo: at(10, 0, 0),
			wantCount: 60,
		},
	}
}

// windowFromTests the windows read from From, and the capped ones.
func windowFromTests() []windowTest {
	at := func(h, m, s int) time.Time {
		return time.Date(2026, 1, 2, h, m, s, 0, time.UTC)
	}

	return []windowTest{
		{
			name:     "from before the span is honoured",
			give:     &Filter{From: at(1, 0, 0), To: at(10, 0, 30)},
			giveSpec: _res1h,
			wantFrom: at(1, 0, 0), wantTo: at(10, 1, 0),
			wantCount: 541,
		},
		{
			name:     "to only",
			give:     &Filter{To: at(10, 0, 30)},
			giveSpec: _res1h,
			wantFrom: at(9, 1, 0), wantTo: at(10, 1, 0),
			wantCount: 60,
		},
		{
			name:     "from rounded down, until now",
			give:     &Filter{From: at(9, 12, 30)},
			giveSpec: _res1h,
			wantFrom: at(9, 12, 0), wantTo: at(12, 31, 0),
			wantCount: 199,
		},
	}
}

// windowEdgeTests the windows From does not move or moves too far.
func windowEdgeTests() []windowTest {
	at := func(h, m, s int) time.Time {
		return time.Date(2026, 1, 2, h, m, s, 0, time.UTC)
	}

	return []windowTest{
		{
			name:     "from after the end falls back to the span",
			give:     &Filter{From: at(13, 0, 0)},
			giveSpec: _res1h,
			wantFrom: at(11, 31, 0), wantTo: at(12, 31, 0),
			wantCount: 60,
		},
		{
			name: "capped to the latest buckets",
			give: &Filter{
				From: time.Date(2025, 12, 30, 12, 0, 0, 0, time.UTC),
			},
			giveSpec: _res1h,
			wantFrom: time.Date(2026, 1, 1, 19, 51, 0, 0, time.UTC),
			wantTo:   at(12, 31, 0), wantCount: MaxBuckets,
			wantTruncated: true,
		},
	}
}

func TestWindowFill(t *testing.T) {
	at := func(millis int64) time.Time { return time.UnixMilli(millis) }
	w := window{from: at(1000), bucket: time.Second, count: 2}
	points := []*Point{
		{Time: at(2000), ResourceId: "r1", Key: 500, Total: 3},
		{Time: at(1000), ResourceId: "r1", Key: 200, Total: 7},
	}

	got := w.fill(points, []string{"r2", "r1"})

	assert.Equal(t, []*Point{
		{Time: at(1000), ResourceId: "r1", Key: 200, Total: 7},
		{Time: at(1000), ResourceId: "r1", Key: 500},
		{Time: at(1000), ResourceId: "r2"},
		{Time: at(2000), ResourceId: "r1", Key: 200},
		{Time: at(2000), ResourceId: "r1", Key: 500, Total: 3},
		{Time: at(2000), ResourceId: "r2"},
	}, got)
}

func TestWindowFillEmpty(t *testing.T) {
	w := window{from: time.UnixMilli(0), bucket: time.Second, count: 3}

	assert.Empty(t, w.fill(nil, nil))
}

func TestFilterResolve(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	filter := &Filter{From: now.Add(-3 * time.Hour)}

	got := filter.Resolve(now)

	// the range is read from its start to the end of the current bucket
	want := Window{
		Interval: _int5m,
		From:     now.Add(-3 * time.Hour),
		To:       now.Add(5 * time.Minute),
	}
	assert.Equal(t, _res3h, filter.Resolution)
	assert.Equal(t, want, got)
	assert.Equal(t, want.From, filter.From)
	assert.Equal(t, want.To, filter.To)
}

func TestFilterResolveKeepsWindow(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	filter := &Filter{From: now.Add(-3 * time.Hour)}
	got := filter.Resolve(now)

	// a later read, in the next bucket, reads the window returned
	_, w, err := planOf(filter, now.Add(5*time.Minute))

	require.NoError(t, err)
	assert.Equal(t, got.From, w.from)
	assert.Equal(t, got.To, w.to())
}

func TestFilterResolveUnknown(t *testing.T) {
	got := (&Filter{Resolution: 99}).Resolve(time.Now())

	assert.Equal(t, Window{}, got)
}

// resolutionTest a case of TestFilterResolution.
type resolutionTest struct {
	name string
	give *Filter
	want typepb.HistogramResolution
}

func TestFilterResolution(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)

	for _, tt := range resolutionTests(now) {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.resolution(now))
		})
	}
}

// resolutionTests the filters of ranges relative to now.
func resolutionTests(now time.Time) []resolutionTest {
	return append([]resolutionTest{
		{name: "nil", give: nil, want: _res1h},
		{name: "set", give: &Filter{Resolution: _res24h}, want: _res24h},
		{name: "unbounded", give: &Filter{}, want: _res1h},
		{
			name: "inverted",
			give: &Filter{From: now, To: now.Add(-time.Minute)},
			want: _res1h,
		},
	}, slices.Concat(rangeResolutionTests(now),
		toleranceResolutionTests(now))...)
}

// rangeResolutionTests the filters of a range picking the Resolution.
func rangeResolutionTests(now time.Time) []resolutionTest {
	return []resolutionTest{
		{
			name: "zoom",
			give: &Filter{
				From: now.Add(-10 * time.Minute),
				To:   now.Add(-5 * time.Minute),
			},
			want: _res5m,
		},
		{
			name: "zoom until now",
			give: &Filter{From: now.Add(-time.Minute)},
			want: _res5m,
		},
		{name: "wide", give: &Filter{From: now.Add(-time.Hour)}, want: _res1h},
		{
			name: "expired seconds",
			give: &Filter{
				From: now.Add(-8 * 24 * time.Hour),
				To:   now.Add(-8*24*time.Hour + time.Minute),
			},
			want: _res30m,
		},
		{
			name: "beyond the longest",
			give: &Filter{From: now.Add(-60 * 24 * time.Hour)},
			want: _res30d,
		},
	}
}

// toleranceResolutionTests the ranges longer than a Resolution: by up to
// one of its buckets they keep it, else take the next one.
func toleranceResolutionTests(now time.Time) []resolutionTest {
	week := 7 * 24 * time.Hour

	return []resolutionTest{
		{
			name: "7 days and a bucket",
			give: &Filter{From: now.Add(-week - 3*time.Hour)},
			want: _res7d,
		},
		{
			name: "7 days and more than a bucket",
			give: &Filter{From: now.Add(-week - 4*time.Hour)},
			want: _res30d,
		},
	}
}
