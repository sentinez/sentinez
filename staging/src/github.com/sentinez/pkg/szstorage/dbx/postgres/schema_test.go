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
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sentinez/pkg/szstorage/dbx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ Client = (*fakeClient)(nil)

// fakeClient records the statements it executes, failing with err.
type fakeClient struct {
	stmts []string
	err   error
}

func (c *fakeClient) Exec(_ context.Context, sql string,
	_ ...any) (pgconn.CommandTag, error) {

	for _, stmt := range strings.Split(sql, ";") {
		c.stmts = append(c.stmts, strings.Join(strings.Fields(stmt), " "))
	}

	return pgconn.CommandTag{}, c.err
}

func (c *fakeClient) Query(context.Context, string,
	...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}

func (c *fakeClient) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

func TestSyncSchema(t *testing.T) {
	client := &fakeClient{}
	schema := dbx.Schema{
		Columns: []dbx.Column{
			{Name: "name", Type: dbx.Text},
			{Name: "content", Type: dbx.JSONB},
		},
		Indexes: []dbx.Index{{Column: "content", Method: dbx.GIN}},
	}

	require.NoError(t, syncSchema(context.Background(), client, "t", schema))

	require.Len(t, client.stmts, 4)
	assert.Contains(t, client.stmts[0], "CREATE TABLE IF NOT EXISTS t (")
	assert.Equal(t, []string{
		"ALTER TABLE t ADD COLUMN IF NOT EXISTS name TEXT",
		"ALTER TABLE t ADD COLUMN IF NOT EXISTS content JSONB",
		"CREATE INDEX IF NOT EXISTS t_content_gin_idx ON t USING GIN (content)",
	}, client.stmts[1:])
}

func TestSyncSchemaError(t *testing.T) {
	client := &fakeClient{err: errors.New("boom")}

	err := syncSchema(context.Background(), client, "t", dbx.Schema{})

	require.ErrorContains(t, err, "sync schema t: boom")
}
