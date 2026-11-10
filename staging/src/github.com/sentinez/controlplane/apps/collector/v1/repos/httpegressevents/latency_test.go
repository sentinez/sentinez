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

package httpegresseventsrepo

import (
	"errors"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatencyQuery(t *testing.T) {
	filter := &eventfilter.Filter{
		ResourceIds: []string{"r1"},
		From:        time.UnixMilli(1000),
		Matches:     []eventfilter.Match{{Field: "status", Value: "502"}},
	}
	b, err := filter.Where(sq.Select().From("t"), _spec.Columns)
	require.NoError(t, err)

	sql, args, err := latencyQuery(b).ToSql()

	require.NoError(t, err)
	assert.Equal(t, "SELECT upstream, count(*), "+
		"coalesce(avg(latency_micros), 0)::BIGINT, "+
		"coalesce(percentile_cont(0.5) WITHIN GROUP "+
		"(ORDER BY latency_micros), 0)::BIGINT, "+
		"coalesce(percentile_cont(0.95) WITHIN GROUP "+
		"(ORDER BY latency_micros), 0)::BIGINT, "+
		"coalesce(percentile_cont(0.99) WITHIN GROUP "+
		"(ORDER BY latency_micros), 0)::BIGINT, "+
		"coalesce(max(latency_micros), 0) FROM t "+
		"WHERE (status) = ? AND resource_id IN (?) AND timestamp >= ? "+
		"GROUP BY upstream ORDER BY upstream", sql)
	assert.Equal(t, []any{int32(502), "r1", int64(1000)}, args)
}

// _rows a dbx.Rows of rows of upstream latencies.
type _rows struct {
	rows []_row
	at   int
	err  error
}

type _row struct {
	upstream string
	total    int64
	micros   [5]int64 // avg, p50, p95, p99, max
}

func (r *_rows) Next() bool {
	r.at++

	return r.at <= len(r.rows)
}

func (r *_rows) Err() error { return r.err }

func (r *_rows) Scan(dest ...any) error {
	row := r.rows[r.at-1]
	*dest[0].(*string) = row.upstream
	*dest[1].(*int64) = row.total
	for i, m := range row.micros {
		*dest[2+i].(*int64) = m
	}

	return nil
}

func TestScanLatency(t *testing.T) {
	rows := &_rows{rows: []_row{
		{"http://a", 3, [5]int64{100, 10, 21, 30, 40}},
		{"http://b", 1, [5]int64{5, 5, 5, 5, 5}},
	}}

	got, err := scanLatency(rows)

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "http://a", got[0].GetUpstream())
	assert.Equal(t, int64(3), got[0].GetTotal())
	assert.Equal(t, int64(100), got[0].GetAvgMicros())
	assert.Equal(t, int64(10), got[0].GetP50Micros())
	assert.Equal(t, int64(21), got[0].GetP95Micros())
	assert.Equal(t, int64(30), got[0].GetP99Micros())
	assert.Equal(t, int64(40), got[0].GetMaxMicros())
	assert.Equal(t, "http://b", got[1].GetUpstream())
}

func TestScanLatencyError(t *testing.T) {
	want := errors.New("boom")

	_, err := scanLatency(&_rows{err: want})

	assert.ErrorIs(t, err, want)
}
