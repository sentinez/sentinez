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
	"errors"
	"fmt"
	"time"

	"github.com/olric-data/olric"
	"github.com/redis/go-redis/v9"
)

// CentralDataChannel the channel centraldata publishes the changes of its
// tables on, and the edges subscribe to.
const CentralDataChannel = "sentinez.centraldata.changes"

// _publishTimeout bounds a publication: a slow cluster must not hold the
// caller back.
const _publishTimeout = 5 * time.Second

// Subscribe listens to channel, on the pub/sub of the cluster, once it is
// ready. The messages are on Channel() of the result, which the caller
// closes. Pub/sub is at-most-once: a subscriber learns of what it missed
// from the payloads (see centraldata's versions), not from the channel.
func (c *Cluster) Subscribe(ctx context.Context,
	channel string) (*redis.PubSub, error) {

	select {
	case <-c.ready:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	_, sub, err := c.subscribe(ctx, channel)

	return sub, err
}

// subscribe opens the pub/sub of the cluster on this node, which receives
// what is published on any node, and listens to channel on it.
func (c *Cluster) subscribe(ctx context.Context,
	channel string) (*olric.PubSub, *redis.PubSub, error) {

	ps, err := c.client.NewPubSub(olric.ToAddress(c.addr))
	if err != nil {
		return nil, nil, fmt.Errorf("cluster: new pubsub: %w", err)
	}

	sub := ps.Subscribe(ctx, channel)
	if _, err := sub.Receive(ctx); err != nil {
		_ = sub.Close()

		return nil, nil, fmt.Errorf("cluster: subscribe %s: %w",
			channel, err)
	}

	return ps, sub, nil
}

// Publisher publishes to the pub/sub of a cluster it is not a member of,
// through the Olric addresses of some of its nodes.
type Publisher struct {
	client *olric.ClusterClient
	pubsub *olric.PubSub
}

// NewPublisher connects to the cluster of the Olric addresses addrs.
func NewPublisher(addrs []string) (*Publisher, error) {
	if len(addrs) == 0 {
		return nil, errors.New("cluster: no olric address")
	}

	client, err := olric.NewClusterClient(addrs)
	if err != nil {
		return nil, fmt.Errorf("cluster: connect %v: %w", addrs, err)
	}

	ps, err := client.NewPubSub()
	if err != nil {
		_ = client.Close(context.Background())

		return nil, fmt.Errorf("cluster: new pubsub: %w", err)
	}

	return &Publisher{client: client, pubsub: ps}, nil
}

// Publish sends msg to the subscribers of channel.
func (p *Publisher) Publish(ctx context.Context, channel string,
	msg []byte) error {

	ctx, cancel := context.WithTimeout(ctx, _publishTimeout)
	defer cancel()

	if _, err := p.pubsub.Publish(ctx, channel, msg); err != nil {
		return fmt.Errorf("cluster: publish %s: %w", channel, err)
	}

	return nil
}

// Close ends the connections.
func (p *Publisher) Close(ctx context.Context) error {
	return p.client.Close(ctx)
}
