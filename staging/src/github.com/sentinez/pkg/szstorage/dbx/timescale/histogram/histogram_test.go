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
	"testing"
	"time"

	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	"github.com/sentinez/pkg/sztable"
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

// viewSchemaTest a case of TestViewSchema.
type viewSchemaTest struct {
	give       Resolution
	wantName   sztable.Table
	wantBucket time.Duration
	wantKeep   time.Duration
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
			assert.Greater(t, view.Refresh.StartOffset,
				2*tt.wantBucket+view.Refresh.EndOffset)
		})
	}
}

// viewSchemaTests the views of a source kept for sourceKeep.
func viewSchemaTests(sourceKeep time.Duration) []viewSchemaTest {
	return []viewSchemaTest{
		{
			give: Minute, wantName: "a.events.minutely",
			wantBucket: time.Minute, wantKeep: sourceKeep,
		},
		{
			give: Hour, wantName: "a.events.hourly",
			wantBucket: time.Hour, wantKeep: sourceKeep,
		},
		{
			give: Second, wantName: "a.events.secondly",
			wantBucket: time.Second, wantKeep: _secondRetention,
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
			view := tt.give.schema(Minute)

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
			wantName:    "a.events.minutely",
			wantGroupBy: []string{"resource_id", "tenant_id"},
		},
		{
			name:        "spec with key",
			give:        Spec{Source: source, Key: "k"}.def(),
			wantName:    "a.events.minutely",
			wantGroupBy: []string{"resource_id", "tenant_id", "k"},
		},
		{
			name: "named",
			give: viewDef{
				source: source, name: "client_ip",
				groupBy: []string{"client_ip"},
			},
			wantName: "a.events.client_ip.minutely",
			wantGroupBy: []string{
				"resource_id", "tenant_id", "client_ip",
			},
		},
	}
}

func TestViewsOf(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 30, 15, 0, time.UTC)
	hour := &timescale.View[Point]{}

	view, w, err := views{Hour: hour}.of(&Filter{Resolution: Hour}, now)

	require.NoError(t, err)
	assert.Same(t, hour, view)
	assert.Equal(t, window{
		from:   time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC),
		bucket: time.Hour,
		count:  24,
	}, w)
}

func TestViewsOfUnknown(t *testing.T) {
	_, _, err := views{}.of(&Filter{Resolution: Hour}, time.Now())

	require.EqualError(t, err, "histogram: unknown resolution 3")
}

// windowTest a case of TestFilterWindow.
type windowTest struct {
	name     string
	give     *Filter
	giveSpec Resolution
	wantFrom time.Time
	wantTo   time.Time
}

func TestFilterWindow(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 30, 15, 0, time.UTC)

	for _, tt := range windowTests() {
		t.Run(tt.name, func(t *testing.T) {
			spec := _resolutions[tt.giveSpec]
			w := tt.give.window(spec, now)

			assert.Equal(t, tt.wantFrom, w.from)
			assert.Equal(t, tt.wantTo, w.to())
			assert.Equal(t, spec.count, w.count)
		})
	}
}

// windowTests the windows ending with the current bucket of 12:30:15,
// before To.
func windowTests() []windowTest {
	at := func(h, m, s int) time.Time {
		return time.Date(2026, 1, 2, h, m, s, 0, time.UTC)
	}

	return []windowTest{
		{
			name: "hour until now", give: nil, giveSpec: Hour,
			wantFrom: at(12, 0, 0).Add(-23 * time.Hour),
			wantTo:   at(13, 0, 0),
		},
		{
			name: "minute until now", give: &Filter{}, giveSpec: Minute,
			wantFrom: at(11, 31, 0), wantTo: at(12, 31, 0),
		},
		{
			name: "second until now", give: &Filter{}, giveSpec: Second,
			wantFrom: at(12, 29, 16), wantTo: at(12, 30, 16),
		},
		{
			name:     "minute until to",
			give:     &Filter{From: at(1, 0, 0), To: at(10, 0, 30)},
			giveSpec: Minute,
			wantFrom: at(9, 1, 0), wantTo: at(10, 1, 0),
		},
		{
			name:     "on a bucket start, to excluded",
			give:     &Filter{To: at(10, 0, 0)},
			giveSpec: Minute,
			wantFrom: at(9, 0, 0), wantTo: at(10, 0, 0),
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

// resolutionTest a case of TestFilterResolution.
type resolutionTest struct {
	name string
	give *Filter
	want Resolution
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
	return []resolutionTest{
		{name: "nil", give: nil, want: Minute},
		{name: "set", give: &Filter{Resolution: Hour}, want: Hour},
		{name: "unbounded", give: &Filter{}, want: Minute},
		{
			name: "zoom",
			give: &Filter{
				From: now.Add(-10 * time.Minute),
				To:   now.Add(-5 * time.Minute),
			},
			want: Second,
		},
		{
			name: "zoom until now",
			give: &Filter{From: now.Add(-time.Minute)},
			want: Second,
		},
		{name: "wide", give: &Filter{From: now.Add(-time.Hour)}, want: Minute},
		{
			name: "expired seconds",
			give: &Filter{
				From: now.Add(-8 * 24 * time.Hour),
				To:   now.Add(-8*24*time.Hour + time.Minute),
			},
			want: Minute,
		},
		{
			name: "inverted",
			give: &Filter{From: now, To: now.Add(-time.Minute)},
			want: Minute,
		},
	}
}
