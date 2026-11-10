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
	"context"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/sentinez/controlplane/apps/collector/v1/repos/eventfilter"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/sztable"
	analyticpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/analytic/v1"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// LatencyReader reads the latency of the origins the requests were
// proxied to.
type LatencyReader interface {
	// Latency measures the proxied requests of filter, but its page, per
	// upstream, ordered by upstream. A match on a field the events do not
	// have is an eventfilter.ErrUnknownField.
	Latency(ctx context.Context,
		filter *eventfilter.Filter) ([]*analyticpb.UpstreamLatency, error)
}

var _ LatencyReader = (*latencyReader)(nil)

// latencyReader the LatencyReader of the egress table.
type latencyReader struct {
	db   dbx.Database[analyticpb.UpstreamLatency]
	cols eventfilter.Columns
}

// openLatency returns the latencyReader of the egress table, which must
// exist (synced by its repo), filtered on cols.
func openLatency(ctx context.Context, conf *settingpb.Config,
	table sztable.Table, cols eventfilter.Columns) (*latencyReader, error) {
	db, err := postgres.Open[analyticpb.UpstreamLatency](ctx, conf,
		dbx.Schema{Table: table}, postgres.NoSync)
	if err != nil {
		return nil, err
	}

	return &latencyReader{db: db, cols: cols}, nil
}

// Latency implements LatencyReader.
func (l *latencyReader) Latency(ctx context.Context,
	filter *eventfilter.Filter) ([]*analyticpb.UpstreamLatency, error) {
	b, err := filter.Where(sq.Select().From(l.db.Table()), l.cols)
	if err != nil {
		return nil, err
	}

	return l.db.CollectRows(ctx, latencyQuery(b), scanLatency)
}

// latencyQuery selects, per upstream of the rows b selects, their number
// and the average, percentiles and maximum of their latency, rounded to a
// microsecond.
func latencyQuery(b sq.SelectBuilder) sq.SelectBuilder {
	const (
		upstream = httppb.EgressEvent_Upstream
		latency  = httppb.EgressEvent_LatencyMicros
	)

	return b.Columns(
		upstream,
		"count(*)",
		"coalesce(avg("+latency+"), 0)::BIGINT",
		percentile(0.5, latency),
		percentile(0.95, latency),
		percentile(0.99, latency),
		"coalesce(max("+latency+"), 0)",
	).GroupBy(upstream).OrderBy(upstream)
}

// percentile the SQL of the p-th percentile (0 to 1) of column.
func percentile(p float64, column string) string {
	return "coalesce(percentile_cont(" + strconv.FormatFloat(p, 'f', -1, 64) +
		") WITHIN GROUP (ORDER BY " + column + "), 0)::BIGINT"
}

// scanLatency scans the rows of a latencyQuery.
func scanLatency(rows dbx.Rows) ([]*analyticpb.UpstreamLatency, error) {
	var latencies []*analyticpb.UpstreamLatency
	for rows.Next() {
		latency := &analyticpb.UpstreamLatency{}

		err := rows.Scan(&latency.Upstream, &latency.Total,
			&latency.AvgMicros, &latency.P50Micros, &latency.P95Micros,
			&latency.P99Micros, &latency.MaxMicros)
		if err != nil {
			return nil, err
		}
		latencies = append(latencies, latency)
	}

	return latencies, rows.Err()
}
