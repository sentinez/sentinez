// Copyright 2026 Sentinéz Labs.
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
	"strconv"
	"testing"
	"time"

	"github.com/olric-data/olric/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startServing starts a node, joining peer if set, and returns its Cluster,
// the address of its Olric server (the one a Publisher dials) and the
// address of its memberlist (the one a peer joins).
func startServing(t *testing.T, peer string) (*Cluster, string, string) {
	t.Helper()

	c := &Cluster{node: "edge", ready: make(chan struct{}),
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

	require.NoError(t, c.open(conf))
	c.Start()
	t.Cleanup(func() { _ = c.Shutdown(context.Background()) })

	<-c.Ready()

	return c, "127.0.0.1:" + strconv.Itoa(conf.BindPort),
		"127.0.0.1:" + strconv.Itoa(conf.MemberlistConfig.BindPort)
}

func TestPublisherReachesSubscriber(t *testing.T) {
	c, addr, _ := startServing(t, "")
	ctx := context.Background()

	sub, err := c.Subscribe(ctx, "test.changes")
	require.NoError(t, err)
	defer func() { _ = sub.Close() }()

	pub, err := NewPublisher([]string{addr})
	require.NoError(t, err)
	defer func() { _ = pub.Close(ctx) }()

	require.NoError(t, pub.Publish(ctx, "test.changes", []byte("hello")))

	select {
	case msg := <-sub.Channel():
		assert.Equal(t, "hello", msg.Payload)
	case <-time.After(_syncWait):
		require.FailNow(t, "the subscriber got nothing")
	}
}

func TestSubscribeWaitsForTheCluster(t *testing.T) {
	c := &Cluster{ready: make(chan struct{})}
	ctx, cancel := context.WithTimeout(context.Background(),
		50*time.Millisecond)
	defer cancel()

	_, err := c.Subscribe(ctx, "test.changes")

	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestNewPublisherNeedsAddress(t *testing.T) {
	_, err := NewPublisher(nil)

	require.Error(t, err)
}

// The edges are a cluster: a change published to any node reaches the
// subscribers of all of them.
func TestPublisherReachesSubscribersOfOtherNodes(t *testing.T) {
	a, addrA, memberA := startServing(t, "")
	b, _, _ := startServing(t, memberA)
	ctx := context.Background()
	require.Eventually(t, func() bool {
		members, err := a.client.Members(ctx)
		return err == nil && len(members) == 2
	}, _syncWait, 50*time.Millisecond)

	// a node that joined needs a moment before it routes: a subscriber
	// retries, as the follower of an edge does.
	subscribe := func(c *Cluster) *redis.PubSub {
		var sub *redis.PubSub
		require.Eventually(t, func() bool {
			var err error
			sub, err = c.Subscribe(ctx, "test.changes")
			return err == nil
		}, _syncWait, 50*time.Millisecond)

		return sub
	}
	subA, subB := subscribe(a), subscribe(b)
	defer func() { _ = subA.Close() }()
	defer func() { _ = subB.Close() }()

	pub, err := NewPublisher([]string{addrA})
	require.NoError(t, err)
	defer func() { _ = pub.Close(ctx) }()
	require.NoError(t, pub.Publish(ctx, "test.changes", []byte("hello")))

	for name, sub := range map[string]*redis.PubSub{"a": subA, "b": subB} {
		select {
		case msg := <-sub.Channel():
			assert.Equal(t, "hello", msg.Payload, name)
		case <-time.After(_syncWait):
			assert.Fail(t, "no message", name)
		}
	}
}
