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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValuesQuery(t *testing.T) {
	for _, tt := range valuesQueryTests() {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := valuesQuery("v", "ip", tt.give, _window).ToSql()

			require.NoError(t, err)
			assert.Equal(t, tt.wantSql, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func valuesQueryTests() []queryTest {
	return []queryTest{
		{
			name: "no filter",
			wantSql: "SELECT bucket, resource_id, ip, " +
				"sum(total)::BIGINT FROM v " +
				"WHERE bucket >= ? AND bucket < ? " +
				"GROUP BY bucket, resource_id, ip " +
				"ORDER BY bucket, resource_id, " +
				"sum(total)::BIGINT DESC, ip",
			wantArgs: []any{int64(1000), int64(3000)},
		},
		{
			name: "filter",
			give: &Filter{TenantId: "t1"},
			wantSql: "SELECT bucket, resource_id, ip, " +
				"sum(total)::BIGINT FROM v " +
				"WHERE bucket >= ? AND bucket < ? AND tenant_id = ? " +
				"GROUP BY bucket, resource_id, ip " +
				"ORDER BY bucket, resource_id, " +
				"sum(total)::BIGINT DESC, ip",
			wantArgs: []any{int64(1000), int64(3000), "t1"},
		},
	}
}

func TestDistinctQuery(t *testing.T) {
	for _, tt := range distinctQueryTests() {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := distinctQuery("v", "ip", tt.give,
				_window).ToSql()

			require.NoError(t, err)
			assert.Equal(t, tt.wantSql, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func distinctQueryTests() []queryTest {
	return []queryTest{
		{
			name: "no filter",
			wantSql: "SELECT bucket, resource_id, 0, " +
				"count(DISTINCT ip)::BIGINT FROM v " +
				"WHERE bucket >= ? AND bucket < ? " +
				"GROUP BY bucket, resource_id " +
				"ORDER BY bucket, resource_id",
			wantArgs: []any{int64(1000), int64(3000)},
		},
		{
			name: "filter",
			give: &Filter{ResourceIds: []string{"r1"}},
			wantSql: "SELECT bucket, resource_id, 0, " +
				"count(DISTINCT ip)::BIGINT FROM v " +
				"WHERE bucket >= ? AND bucket < ? AND resource_id IN (?) " +
				"GROUP BY bucket, resource_id " +
				"ORDER BY bucket, resource_id",
			wantArgs: []any{int64(1000), int64(3000), "r1"},
		},
	}
}
