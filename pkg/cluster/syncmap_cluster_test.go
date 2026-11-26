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
	"net"
	"testing"
	"time"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _syncWait = 10 * time.Second

type serverMap = SyncMap[edgepb.Resource, *edgepb.Resource]

func freePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	return l.Addr().(*net.TCPAddr).Port
}

type serverHook = func(key string, prev, next *edgepb.Resource)

// startPair starts two joined nodes and a started map name on each;
// hook, if not nil, is the OnChange of the map of the second node.
func startPair(
	t *testing.T, name string, hook serverHook,
) (*serverMap, *serverMap) {
	t.Helper()

	a, _, member := startNodeWith(t, "edge-a", "", nil)
	b, _, _ := startNodeWith(t, "edge-b", member, nil)
	waitMembers(t, a, 2)

	return startMap(t, a, name, nil), startMap(t, b, name, hook)
}

func startMap(
	t *testing.T, c *Cluster, name string, hook serverHook,
) *serverMap {
	t.Helper()

	m := NewSyncMap[edgepb.Resource](c, name)
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

	require.NoError(t, a.Store(ctx, "k", &edgepb.Resource{Name: "v1"}))
	require.Eventually(t, loaded(b, "k", "v1"), _syncWait, 10*time.Millisecond)

	require.NoError(t, b.Delete(ctx, "k"))
	require.Eventually(t, missing(a, "k"), _syncWait, 10*time.Millisecond)
}

func TestSyncMapClusterLateJoinerLoads(t *testing.T) {
	ctx := context.Background()
	a, b := startPair(t, "late", nil)
	require.NoError(t, a.Store(ctx, "k", &edgepb.Resource{Name: "v1"}))
	require.Eventually(t, loaded(b, "k", "v1"), _syncWait, 10*time.Millisecond)

	late := startMap(t, b.cluster, "late", nil)

	assert.True(t, loaded(late, "k", "v1")(), "loaded by Start")
}

func TestSyncMapClusterRepairsLostMessage(t *testing.T) {
	a, b := startPair(t, "repair", nil)

	// a write whose PubSub message is lost: only the local map has it
	e := entry[edgepb.Resource]{value: &edgepb.Resource{Name: "v1"},
		record: record{version: a.clock.next(), origin: "edge-a"}}
	e.payload = mustMarshal(t, e.value)
	a.apply("k", e)

	require.Eventually(t, loaded(b, "k", "v1"), _syncWait, 10*time.Millisecond)
}

func TestSyncMapClusterOnChangeFromPeer(t *testing.T) {
	changed := make(chan string, 1)
	a, _ := startPair(t, "on-change",
		func(_ string, _, next *edgepb.Resource) { changed <- next.GetName() })

	require.NoError(t, a.Store(context.Background(), "k",
		&edgepb.Resource{Name: "v1"}))

	select {
	case got := <-changed:
		assert.Equal(t, "v1", got)
	case <-time.After(_syncWait):
		t.Fatal("no change seen on the peer")
	}
}
