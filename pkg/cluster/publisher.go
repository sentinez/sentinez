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
	"crypto/tls"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/olric-data/olric"
	"github.com/olric-data/olric/config"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/sentinez/pkg/cluster/channel"
)

// _publishTimeout bounds one attempt of a publication: a slow cluster must
// not hold the caller back.
const _publishTimeout = 5 * time.Second

var errNoAddress = errors.New("cluster: no olric address")

// Publisher publishes to the pub/sub of a cluster it is not a member of,
// through the Olric addresses of some of its nodes, the seeds. A node
// relays what it is published to the subscribers of every node, so one
// node is enough: the publisher holds the pub/sub of one seed, and moves
// to the next one when a publication fails.
type Publisher struct {
	client *olric.ClusterClient
	seeds  []string

	mu sync.Mutex
	// pubsub is on seeds[at]
	pubsub *olric.PubSub
	at     int
}

// PublisherOption configures NewPublisher.
type PublisherOption interface {
	apply(opts *[]olric.ClusterClientOption)
}

type tlsOption struct{ conf *tls.Config }

func (t tlsOption) apply(opts *[]olric.ClusterClientOption) {
	if t.conf == nil {
		return
	}

	conf := config.NewClient()
	conf.TLSConfig = t.conf
	*opts = append(*opts, olric.WithConfig(conf))
}

// WithTLS dials the nodes with TLS: Olric does not serve TLS, so a TLS
// proxy must stand in front of each node. nil is plaintext.
func WithTLS(conf *tls.Config) PublisherOption {
	return tlsOption{conf: conf}
}

type passwordOption string

func (p passwordOption) apply(opts *[]olric.ClusterClientOption) {
	if p != "" {
		*opts = append(*opts, olric.WithPassword(string(p)))
	}
}

// WithPassword authenticates to the nodes with password
// (SZ_OLRIC_PASSWORD); empty does not authenticate.
func WithPassword(password string) PublisherOption {
	return passwordOption(password)
}

// NewPublisher connects to the cluster through the first of the Olric
// addresses seeds that answers: a dead seed is skipped, it fails only when
// none answers.
func NewPublisher(seeds []string, opts ...PublisherOption) (*Publisher,
	error) {
	if len(seeds) == 0 {
		return nil, errNoAddress
	}

	clientOpts := clientOptions(opts)

	var errs []error
	for at, seed := range seeds {
		// one seed per client: given them all, the client asks a random
		// one for the members, and fails when it is a dead one
		client, err := olric.NewClusterClient([]string{seed}, clientOpts...)
		if err != nil {
			errs = append(errs, fmt.Errorf("connect %s: %w", seed, err))
			continue
		}

		ps, err := pubsubOn(client, seed)
		if err != nil {
			_ = client.Close(context.Background())
			errs = append(errs, err)

			continue
		}

		return &Publisher{client: client, seeds: seeds, pubsub: ps,
			at: at}, nil
	}

	return nil, fmt.Errorf("cluster: no seed answers: %w",
		errors.Join(errs...))
}

// Publish sends msg to the subscribers of ch. When it fails, it moves to
// the next seed and tries once more.
func (p *Publisher) Publish(ctx context.Context, ch channel.Channel,
	msg []byte) error {
	ps, at := p.current()

	err := publish(ctx, ps, ch, msg)
	if err == nil || ctx.Err() != nil {
		return err
	}

	next, nextAt, nextErr := p.moveOn(at)
	if nextErr != nil {
		return errors.Join(err, nextErr)
	}

	szlog.Warnf("%v, retry on %s", err, p.seeds[nextAt])

	return publish(ctx, next, ch, msg)
}

// Close ends the connections.
func (p *Publisher) Close(ctx context.Context) error {
	return p.client.Close(ctx)
}

// current the pub/sub publications go through, and the index of its seed.
func (p *Publisher) current() (*olric.PubSub, int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.pubsub, p.at
}

// moveOn puts the pub/sub on the seed after failed, the index of the seed
// that failed, and returns it with the index of its seed; when another
// publication already moved, it keeps that one.
func (p *Publisher) moveOn(failed int) (*olric.PubSub, int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.at != failed {
		return p.pubsub, p.at, nil
	}

	at := (failed + 1) % len(p.seeds)

	ps, err := pubsubOn(p.client, p.seeds[at])
	if err != nil {
		return nil, 0, err
	}

	p.pubsub, p.at = ps, at

	return ps, at, nil
}

// pubsubOn the pub/sub of client on the node of seed.
func pubsubOn(client *olric.ClusterClient, seed string) (*olric.PubSub,
	error) {
	ps, err := client.NewPubSub(olric.ToAddress(seed))
	if err != nil {
		return nil, fmt.Errorf("cluster: pubsub on %s: %w", seed, err)
	}

	return ps, nil
}

// publish sends msg on ps, within _publishTimeout.
func publish(ctx context.Context, ps *olric.PubSub, ch channel.Channel,
	msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, _publishTimeout)
	defer cancel()

	if _, err := ps.Publish(ctx, ch.String(), msg); err != nil {
		return fmt.Errorf("cluster: publish %s: %w", ch, err)
	}

	return nil
}

// clientOptions the options of the cluster client of opts.
func clientOptions(opts []PublisherOption) []olric.ClusterClientOption {
	var clientOpts []olric.ClusterClientOption
	for _, opt := range opts {
		opt.apply(&clientOpts)
	}

	return clientOpts
}
