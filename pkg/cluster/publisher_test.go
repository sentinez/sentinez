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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deadAddr an address nothing listens on.
func deadAddr(t *testing.T) string {
	t.Helper()

	return "127.0.0.1:" + strconv.Itoa(freePort(t))
}

func TestNewPublisherSkipsDeadSeeds(t *testing.T) {
	c, addr, _ := startServing(t, "")
	sub := subscribeTo(t, c)
	ctx := context.Background()

	pub, err := NewPublisher([]string{deadAddr(t), deadAddr(t), addr})
	require.NoError(t, err)
	defer func() { _ = pub.Close(ctx) }()

	assert.Equal(t, 2, pub.at)
	require.NoError(t, pub.Publish(ctx, "test.changes", []byte("hello")))
	assert.Equal(t, "hello", receive(t, sub))
}

func TestNewPublisherWithoutLiveSeed(t *testing.T) {
	_, err := NewPublisher([]string{deadAddr(t), deadAddr(t)})

	require.Error(t, err)
}

// The node the publisher holds dies: the publication moves to the next
// seed, and the subscribers of the live node still get it.
func TestPublishMovesToTheNextSeed(t *testing.T) {
	a, addrA, memberA := startServing(t, "")
	b, addrB, _ := startServing(t, memberA)
	waitMembers(t, b, 2)
	sub := subscribeTo(t, b)
	ctx := context.Background()

	pub, err := NewPublisher([]string{addrA, addrB})
	require.NoError(t, err)
	defer func() { _ = pub.Close(ctx) }()
	require.Equal(t, 0, pub.at)

	require.NoError(t, a.Shutdown(ctx))

	require.NoError(t, pub.Publish(ctx, "test.changes", []byte("hello")))
	assert.Equal(t, 1, pub.at)
	assert.Equal(t, "hello", receive(t, sub))
}

// With every seed dead, a publication fails after one retry.
func TestPublishFailsWhenEverySeedIsDead(t *testing.T) {
	a, addrA, _ := startServing(t, "")
	ctx := context.Background()

	pub, err := NewPublisher([]string{addrA})
	require.NoError(t, err)
	defer func() { _ = pub.Close(ctx) }()

	require.NoError(t, a.Shutdown(ctx))

	require.Error(t, pub.Publish(ctx, "test.changes", []byte("hello")))
}
