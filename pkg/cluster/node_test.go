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
	"strconv"
	"testing"
	"time"

	"github.com/olric-data/olric/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// startServing starts a node, joining peer if set, and returns its Cluster,
// the address of its Olric server (the one a Publisher dials) and the
// address of its memberlist (the one a peer joins).
func startServing(t *testing.T, peer string) (*Cluster, string, string) {
	t.Helper()

	return startNodeWith(t, "edge", peer, nil)
}

// startNodeWith is startServing for the node named node, its config set by
// set if not nil.
func startNodeWith(t *testing.T, node, peer string,
	set func(*config.Config)) (*Cluster, string, string) {
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
	if set != nil {
		set(conf)
	}

	require.NoError(t, c.open(conf))
	c.Start()
	t.Cleanup(func() { _ = c.Shutdown(context.Background()) })

	<-c.Ready()

	return c, "127.0.0.1:" + strconv.Itoa(conf.BindPort),
		"127.0.0.1:" + strconv.Itoa(conf.MemberlistConfig.BindPort)
}

// waitMembers waits until c sees n members.
func waitMembers(t *testing.T, c *Cluster, n int) {
	t.Helper()

	require.Eventually(t, func() bool {
		members, err := c.client.Members(context.Background())
		return err == nil && len(members) == n
	}, _syncWait, 50*time.Millisecond)
}

// subscribeTo subscribes to test.changes on c, retrying while a node that
// joined is not routing yet, as the follower of an edge does.
func subscribeTo(t *testing.T, c *Cluster) *redis.PubSub {
	t.Helper()

	var sub *redis.PubSub
	require.Eventually(t, func() bool {
		var err error
		sub, err = c.Subscribe(context.Background(), "test.changes")
		return err == nil
	}, _syncWait, 50*time.Millisecond)
	t.Cleanup(func() { _ = sub.Close() })

	return sub
}

// receive the next message of sub.
func receive(t *testing.T, sub *redis.PubSub) string {
	t.Helper()

	select {
	case msg := <-sub.Channel():
		return msg.Payload
	case <-time.After(_syncWait):
		require.FailNow(t, "the subscriber got nothing")
		return ""
	}
}
