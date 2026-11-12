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

package centraldatafac

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	memrepo "github.com/sentinez/controlplane/apps/centraldata/v1/repos/memory"
	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	"github.com/sentinez/pkg/sztable"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chanPublisher hands the published changes to the test.
type chanPublisher struct {
	changes chan *centraldatapb.Change
}

func (p chanPublisher) Publish(ctx context.Context,
	c *centraldatapb.Change) error {

	select {
	case p.changes <- c:
	case <-ctx.Done():
	}

	return nil
}

func next(t *testing.T, p chanPublisher) *centraldatapb.Change {
	t.Helper()

	select {
	case c := <-p.changes:
		return c
	case <-time.After(10 * time.Second):
		require.FailNow(t, "no change was published")

		return nil
	}
}

// liveEnv a started Syncer on a real database, the one of SZ_TEST_PG_URI.
type liveEnv struct {
	conn     *pgx.Conn
	relation string
	service  *centraldatasvc.CentralDataService
	pub      chanPublisher
}

// startLive skips the test without SZ_TEST_PG_URI. The table of the test is
// dropped when it ends.
func startLive(t *testing.T) *liveEnv {
	t.Helper()

	uri := os.Getenv("SZ_TEST_PG_URI")
	if uri == "" {
		t.Skip("SZ_TEST_PG_URI is not set")
	}

	ctx := context.Background()
	conf := liveConf(uri)

	relation, err := postgres.TableName(conf, sztable.SecuritySecRules)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, uri)
	require.NoError(t, err)
	drop := "DROP TABLE IF EXISTS " + relation
	_, err = conn.Exec(ctx, drop)
	require.NoError(t, err)

	service, err := NewDefaultService(conf)
	require.NoError(t, err)
	pub := chanPublisher{changes: make(chan *centraldatapb.Change)}
	syncer, err := NewSyncer(conf, service, pub)
	require.NoError(t, err)
	require.NoError(t, syncer.Start(ctx))

	t.Cleanup(func() {
		syncer.Close()
		_, _ = conn.Exec(ctx, drop)
		_ = conn.Close(ctx)
	})

	return &liveEnv{conn: conn, relation: relation, service: service,
		pub: pub}
}

func liveConf(uri string) *settingpb.Config {
	return &settingpb.Config{
		Flag: &settingpb.Flag{EnvMode: "dev"},
		Env: map[string]string{
			"SENZ_POSTGRES_URI": uri,
			"SENZ_SYNC_TABLES":  "security.secrules",
		},
	}
}

func (e *liveEnv) exec(t *testing.T, sql string) {
	t.Helper()

	_, err := e.conn.Exec(context.Background(), sql)
	require.NoError(t, err)
}

func (e *liveEnv) name(t *testing.T) string {
	t.Helper()

	row, err := memrepo.Get[securitypb.SecRule](e.service.Memory(),
		sztable.SecuritySecRules, "sr-1")
	require.NoError(t, err)

	return row.GetName()
}

// TestSyncLive runs the whole path against a real database: the changes of
// a row reach the memory and the publisher, in order.
func TestSyncLive(t *testing.T) {
	env := startLive(t)

	// every column set: the repositories never write a NULL.
	env.exec(t, "INSERT INTO "+env.relation+
		" (id, name, description, expr, action, status, priority,"+
		" owner_resource_id, owner_tenant_id, owner_created_at,"+
		" owner_updated_at) VALUES ('sr-1', 'first', '', '{}', '', 1, 1,"+
		" 'rs-1', 'tn-1', now(), now())")
	inserted := next(t, env.pub)
	assert.Equal(t, centraldatapb.Change_OP_UPSERT, inserted.GetOp())
	assert.Equal(t, "sr-1", inserted.GetId())
	assert.Equal(t, "first", env.name(t))

	env.exec(t, "UPDATE "+env.relation+" SET name = 'second'")
	assert.Equal(t, inserted.GetVersion()+1, next(t, env.pub).GetVersion())
	assert.Equal(t, "second", env.name(t))

	env.exec(t, "DELETE FROM "+env.relation)
	assert.Equal(t, centraldatapb.Change_OP_DELETE, next(t, env.pub).GetOp())
	_, err := memrepo.Get[securitypb.SecRule](env.service.Memory(),
		sztable.SecuritySecRules, "sr-1")
	require.Error(t, err)
}
