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
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// liveConn a connection to the database of SZ_TEST_PG_URI with a table
// that notifies sz_live_changes; the test is skipped without it.
func liveConn(ctx context.Context, t *testing.T) *pgx.Conn {
	t.Helper()

	uri := os.Getenv("SZ_TEST_PG_URI")
	if uri == "" {
		t.Skip("SZ_TEST_PG_URI is not set")
	}

	conn, err := pgx.Connect(ctx, uri)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = conn.Exec(context.Background(), `DROP TABLE sz_live_notify`)
		_ = conn.Close(context.Background())
	})

	_, err = conn.Exec(ctx, `DROP TABLE IF EXISTS sz_live_notify;
		CREATE TABLE sz_live_notify (id TEXT PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)

	// twice: it must be idempotent.
	for range 2 {
		require.NoError(t, NotifyChange(ctx, conn, "sz_live_notify",
			"sz_live_changes"))
	}

	return conn
}

// TestNotifyChangeLive runs the trigger against a real database, the one
// of SZ_TEST_PG_URI; it is skipped without it.
func TestNotifyChangeLive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	conn := liveConn(ctx, t)
	_, err := conn.Exec(ctx, `LISTEN sz_live_changes`)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `
		INSERT INTO sz_live_notify VALUES ('a', 'x');
		UPDATE sz_live_notify SET name = 'y' WHERE id = 'a';
		DELETE FROM sz_live_notify WHERE id = 'a'`)
	require.NoError(t, err)

	var got []Change
	for range 3 {
		n, err := conn.WaitForNotification(ctx)
		require.NoError(t, err)

		var c Change
		require.NoError(t, json.Unmarshal([]byte(n.Payload), &c))
		got = append(got, c)
	}

	assert.Equal(t, []Change{
		{Table: "sz_live_notify", Op: "INSERT", Id: "a"},
		{Table: "sz_live_notify", Op: "UPDATE", Id: "a"},
		{Table: "sz_live_notify", Op: "DELETE", Id: "a"},
	}, got)
}
