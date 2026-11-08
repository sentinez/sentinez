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

package postgres

import (
	"testing"

	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTable only names its table.
type fakeTable struct {
	dbx.Database[struct{}]
}

func (fakeTable) Table() string { return "items" }

func TestUpsertBuilder(t *testing.T) {
	q := UpsertBuilder[struct{}](fakeTable{}, M{
		"name": "a",
		"id":   "itm_1",
		"plan": 2,
	})

	stmt, args, err := q.ToSql()

	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO items (id,name,plan) VALUES ($1,$2,$3) "+
		"ON CONFLICT (id) DO UPDATE SET "+
		"updated_at = now() AT TIME ZONE 'UTC', "+
		"name = EXCLUDED.name, plan = EXCLUDED.plan RETURNING id", stmt)
	assert.Equal(t, []any{"itm_1", "a", 2}, args)
}
