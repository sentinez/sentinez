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

package cluster

import (
	"context"
	"io"
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/olric-data/olric"
	"github.com/olric-data/olric/config"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _syncWait = 10 * time.Second

type serverMap = SyncMap[edgepb.Server, *edgepb.Server]

func freePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	return l.Addr().(*net.TCPAddr).Port
}

// startNode starts an Olric node on free ports, joining peer if set,
// and returns its Cluster and memberlist address.
func startNode(t *testing.T, node, peer string) (*Cluster, string) {
	t.Helper()

	c := &Cluster{node: node, ready: make(chan struct{}),
		done: make(chan struct{})}

	conf := config.New("local")
	conf.BindAddr = "127.0.0.1"
	conf.BindPort = freePort(t)
	conf.MemberlistConfig.BindAddr = "127.0.0.1"
	conf.MemberlistConfig.BindPort = freePort(t)
	conf.Logger = log.New(io.Discard, "", 0)
	conf.LogOutput = io.Discard
	conf.Started = c.markReady
	if peer != "" {
		conf.Peers = []string{peer}
	}

	db, err := olric.New(conf)
	require.NoError(t, err)
	c.db, c.client = db, db.NewEmbeddedClient()
	c.Start()
	t.Cleanup(func() { _ = c.Shutdown(context.Background()) })

	<-c.Ready()

	return c, "127.0.0.1:" + strconv.Itoa(conf.MemberlistConfig.BindPort)
}

type serverHook = func(key string, prev, next *edgepb.Server)

// startPair starts two joined nodes and a started map name on each;
// hook, if not nil, is the OnChange of the map of the second node.
func startPair(
	t *testing.T, name string, hook serverHook,
) (*serverMap, *serverMap) {
	t.Helper()

	a, addr := startNode(t, "edge-a", "")
	b, _ := startNode(t, "edge-b", addr)
	require.Eventually(t, func() bool {
		members, err := a.client.Members(context.Background())
		return err == nil && len(members) == 2
	}, _syncWait, 50*time.Millisecond)

	return startMap(t, a, name, nil), startMap(t, b, name, hook)
}

func startMap(
	t *testing.T, c *Cluster, name string, hook serverHook,
) *serverMap {
	t.Helper()

	m := NewSyncMap[edgepb.Server](c, name)
	m.interval = 50 * time.Millisecond
	m.OnChange(hook)
	require.NoError(t, m.Start(context.Background()))
	t.Cleanup(func() { _ = m.Close() })

	return m
}

func missing(m *serverMap, key string) func() bool {
	return func() bool {
		_, ok := m.Load(key)
		return !ok
	}
}

func loaded(m *serverMap, key, name string) func() bool {
	return func() bool {
		got, ok := m.Load(key)
		return ok && got.GetName() == name
	}
}

func TestSyncMapClusterStoreDelete(t *testing.T) {
	ctx := context.Background()
	a, b := startPair(t, "store-delete", nil)

	require.NoError(t, a.Store(ctx, "k", &edgepb.Server{Name: "v1"}))
	require.Eventually(t, loaded(b, "k", "v1"), _syncWait, 10*time.Millisecond)

	require.NoError(t, b.Delete(ctx, "k"))
	require.Eventually(t, missing(a, "k"), _syncWait, 10*time.Millisecond)
}

func TestSyncMapClusterLateJoinerLoads(t *testing.T) {
	ctx := context.Background()
	a, b := startPair(t, "late", nil)
	require.NoError(t, a.Store(ctx, "k", &edgepb.Server{Name: "v1"}))
	require.Eventually(t, loaded(b, "k", "v1"), _syncWait, 10*time.Millisecond)

	late := startMap(t, b.cluster, "late", nil)

	assert.True(t, loaded(late, "k", "v1")(), "loaded by Start")
}

func TestSyncMapClusterRepairsLostMessage(t *testing.T) {
	a, b := startPair(t, "repair", nil)

	// a write whose PubSub message is lost: only the local map has it
	e := entry[edgepb.Server]{value: &edgepb.Server{Name: "v1"},
		record: record{version: a.clock.next(), origin: "edge-a"}}
	e.payload = mustMarshal(t, e.value)
	a.apply("k", e)

	require.Eventually(t, loaded(b, "k", "v1"), _syncWait, 10*time.Millisecond)
}

func TestSyncMapClusterOnChangeFromPeer(t *testing.T) {
	changed := make(chan string, 1)
	a, _ := startPair(t, "on-change",
		func(_ string, _, next *edgepb.Server) { changed <- next.GetName() })

	require.NoError(t, a.Store(context.Background(), "k",
		&edgepb.Server{Name: "v1"}))

	select {
	case got := <-changed:
		assert.Equal(t, "v1", got)
	case <-time.After(_syncWait):
		t.Fatal("no change seen on the peer")
	}
}
