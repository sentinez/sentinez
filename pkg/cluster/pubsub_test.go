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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublisherReachesSubscriber(t *testing.T) {
	c, addr, _ := startServing(t, "")
	sub := subscribeTo(t, c)
	ctx := context.Background()

	pub, err := NewPublisher([]string{addr})
	require.NoError(t, err)
	defer func() { _ = pub.Close(ctx) }()

	require.NoError(t, pub.Publish(ctx, "test.changes", []byte("hello")))
	assert.Equal(t, "hello", receive(t, sub))
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
	waitMembers(t, a, 2)
	subA, subB := subscribeTo(t, a), subscribeTo(t, b)
	ctx := context.Background()

	pub, err := NewPublisher([]string{addrA})
	require.NoError(t, err)
	defer func() { _ = pub.Close(ctx) }()
	require.NoError(t, pub.Publish(ctx, "test.changes", []byte("hello")))

	assert.Equal(t, "hello", receive(t, subA), "a")
	assert.Equal(t, "hello", receive(t, subB), "b")
}
