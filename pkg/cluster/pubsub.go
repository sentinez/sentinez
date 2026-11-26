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
	"fmt"

	"github.com/olric-data/olric"
	"github.com/redis/go-redis/v9"
	"github.com/sentinez/sentinez/pkg/cluster/channel"
)

// Subscribe listens to ch, on the pub/sub of the cluster, once it is
// ready. The messages are on Channel() of the result, which the caller
// closes. Pub/sub is at-most-once: a subscriber learns of what it missed
// from the payloads (see centraldata's versions), not from the channel.
func (c *Cluster) Subscribe(ctx context.Context,
	ch channel.Channel) (*redis.PubSub, error) {

	select {
	case <-c.ready:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	_, sub, err := c.subscribe(ctx, ch)

	return sub, err
}

// subscribe opens the pub/sub of the cluster on this node, which receives
// what is published on any node, and listens to ch on it.
func (c *Cluster) subscribe(ctx context.Context,
	ch channel.Channel) (*olric.PubSub, *redis.PubSub, error) {

	ps, err := c.client.NewPubSub(olric.ToAddress(c.addr))
	if err != nil {
		return nil, nil, fmt.Errorf("cluster: new pubsub: %w", err)
	}

	sub := ps.Subscribe(ctx, ch.String())
	if _, err := sub.Receive(ctx); err != nil {
		_ = sub.Close()

		return nil, nil, fmt.Errorf("cluster: subscribe %s: %w",
			ch, err)
	}

	return ps, sub, nil
}
