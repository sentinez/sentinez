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

	"github.com/sentinez/controlplane/pkg/schema"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/sentinez/pkg/szstorage/dbx/timescale"
	"github.com/sentinez/pkg/sztable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPointsQuery(t *testing.T) {
	from := time.UnixMilli(1000)
	to := time.UnixMilli(2000)

	tests := []struct {
		name     string
		giveKey  string
		give     *Filter
		wantSql  string
		wantArgs []any
	}{
		{
			name: "no filter, no key",
			wantSql: "SELECT bucket, resource_id, 0, sum(total)::BIGINT " +
				"FROM v GROUP BY bucket, resource_id " +
				"ORDER BY bucket, resource_id",
		},
		{
			name:    "filter and key",
			giveKey: "status",
			give: &Filter{
				ResourceIds: []string{"r1", "r2"},
				TenantId:    "t1",
				From:        from,
				To:          to,
			},
			wantSql: "SELECT bucket, resource_id, status, " +
				"sum(total)::BIGINT FROM v " +
				"WHERE resource_id IN (?,?) AND tenant_id = ? " +
				"AND bucket >= ? AND bucket < ? " +
				"GROUP BY bucket, resource_id, status " +
				"ORDER BY bucket, resource_id",
			wantArgs: []any{"r1", "r2", "t1", int64(1000), int64(2000)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := pointsQuery("v", tt.giveKey, tt.give).ToSql()

			require.NoError(t, err)
			assert.Equal(t, tt.wantSql, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestViewSchema(t *testing.T) {
	spec := Spec{
		Source: timescale.Hypertable{
			Schema:     dbx.Schema{Table: "a.events"},
			TimeColumn: "ts",
		},
		Key: "status",
	}

	tests := []struct {
		give       int
		wantName   sztable.Table
		wantBucket time.Duration
		wantKeep   time.Duration
	}{
		{
			give: 0, wantName: "a.events.minutely",
			wantBucket: time.Minute, wantKeep: schema.EventRetention,
		},
		{
			give: 1, wantName: "a.events.hourly",
			wantBucket: time.Hour, wantKeep: schema.EventRetention,
		},
		{
			give: 2, wantName: "a.events.secondly",
			wantBucket: time.Second, wantKeep: _secondRetention,
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantName.String(), func(t *testing.T) {
			view := viewSchema(spec, tt.give)

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

func TestFilterResolution(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		give *Filter
		want Resolution
	}{
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
		{
			name: "wide",
			give: &Filter{From: now.Add(-time.Hour)},
			want: Minute,
		},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.resolution(now))
		})
	}
}
