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
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func newLocalSyncMap(node string) *SyncMap[edgepb.Server, *edgepb.Server] {
	return NewSyncMap[edgepb.Server](&Cluster{node: node}, "test")
}

func TestSyncMapLocalBeforeStart(t *testing.T) {
	ctx := context.Background()
	m := newLocalSyncMap("edge-a")

	require.NoError(t, m.Store(ctx, "k", &edgepb.Server{Name: "a"}))
	got, ok := m.Load("k")
	require.True(t, ok)
	assert.Equal(t, "a", got.GetName())

	require.NoError(t, m.Delete(ctx, "k"))
	_, ok = m.Load("k")
	assert.False(t, ok)
}

func TestSyncMapApplyLastWriterWins(t *testing.T) {
	m := newLocalSyncMap("edge-a")
	value := func(name string, version uint64) entry[edgepb.Server] {
		return entry[edgepb.Server]{
			record: record{version: version, origin: "edge-b"},
			value:  &edgepb.Server{Name: name},
		}
	}

	assert.True(t, m.apply("k", value("new", 10)))
	assert.False(t, m.apply("k", value("old", 5)))

	got, _ := m.Load("k")
	assert.Equal(t, "new", got.GetName())
	assert.Greater(t, m.clock.next(), uint64(10), "the clock saw 10")
}

func TestSyncMapRangeSkipsTombstones(t *testing.T) {
	ctx := context.Background()
	m := newLocalSyncMap("edge-a")
	require.NoError(t, m.Store(ctx, "a", &edgepb.Server{Name: "a"}))
	require.NoError(t, m.Store(ctx, "b", &edgepb.Server{Name: "b"}))
	require.NoError(t, m.Delete(ctx, "b"))

	var keys []string
	m.Range(func(key string, _ *edgepb.Server) bool {
		keys = append(keys, key)
		return true
	})

	assert.Equal(t, []string{"a"}, keys)
}

func TestSyncMapDecode(t *testing.T) {
	m := newLocalSyncMap("edge-a")
	payload, err := proto.Marshal(&edgepb.Server{Name: "x"})
	require.NoError(t, err)

	e, err := m.decode(record{version: 1, payload: payload})

	require.NoError(t, err)
	assert.Equal(t, "x", e.value.GetName())
}

func mustMarshal(t *testing.T, m proto.Message) []byte {
	t.Helper()

	raw, err := proto.Marshal(m)
	require.NoError(t, err)

	return raw
}

func TestSyncMapOnChange(t *testing.T) {
	ctx := context.Background()
	m := newLocalSyncMap("edge-a")

	type change struct{ prev, next string }
	var got []change
	m.OnChange(func(_ string, prev, next *edgepb.Server) {
		got = append(got, change{prev.GetName(), next.GetName()})
	})

	require.NoError(t, m.Store(ctx, "k", &edgepb.Server{Name: "v1"}))
	require.NoError(t, m.Store(ctx, "k", &edgepb.Server{Name: "v2"}))
	require.NoError(t, m.Delete(ctx, "k"))
	m.apply("k", entry[edgepb.Server]{record: record{version: 1}})

	assert.Equal(t, []change{{"", "v1"}, {"v1", "v2"}, {"v2", ""}}, got,
		"an older write changes nothing")
}
