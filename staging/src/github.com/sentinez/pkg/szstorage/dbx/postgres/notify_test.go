// Copyright 2025 Duc-Hung Ho.
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
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rawClient keeps the SQL as sent: fakeClient splits it at every ";",
// which the body of the trigger function holds.
type rawClient struct {
	fakeClient
	sql string
}

func (c *rawClient) Exec(ctx context.Context, sql string,
	args ...any) (pgconn.CommandTag, error) {

	c.sql = sql

	return c.fakeClient.Exec(ctx, sql, args...)
}

func TestNotifyChange(t *testing.T) {
	client := &rawClient{}

	err := NotifyChange(context.Background(), client,
		"security_secrules", "sz_changes")

	require.NoError(t, err)
	assert.Contains(t, client.sql, "CREATE OR REPLACE FUNCTION "+
		NotifyFunction+"() RETURNS trigger")
	assert.Contains(t, client.sql, "$$ BEGIN PERFORM pg_notify(TG_ARGV[0],")
	assert.Contains(t, client.sql,
		";DROP TRIGGER IF EXISTS security_secrules_notify ON "+
			"security_secrules;")
	assert.Contains(t, client.sql,
		"CREATE TRIGGER security_secrules_notify AFTER INSERT OR UPDATE "+
			"OR DELETE ON security_secrules FOR EACH ROW EXECUTE "+
			"FUNCTION "+NotifyFunction+"('sz_changes')")
}

func TestNotifyChangeRejectsInjection(t *testing.T) {
	tests := []struct {
		name        string
		giveTable   string
		giveChannel string
	}{
		{"table", "t; DROP TABLE users", "ch"},
		{"channel", "t", "ch'); DROP TABLE users; --"},
		{"empty table", "", "ch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeClient{}

			err := NotifyChange(context.Background(), client,
				tt.giveTable, tt.giveChannel)

			require.Error(t, err)
			assert.Empty(t, client.stmts, "nothing must run")
		})
	}
}

func TestNotifyChangeError(t *testing.T) {
	client := &fakeClient{err: errors.New("boom")}

	err := NotifyChange(context.Background(), client, "t", "ch")

	require.ErrorContains(t, err, "boom")
}
