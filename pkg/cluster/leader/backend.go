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

package leader

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sentinez/sentinez/api/client/consul"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/pkg/cluster"
	"github.com/sentinez/shared/szlog"
)

const (
	_keyPrefix  = "sentinez/dmz/edge/leader/"
	_sessionTTL = 15 * time.Second
)

var (
	errNoConsul = errors.New(
		"leader: clustered edge needs SENZ_CONSUL_URI to elect a leader")
	errNoZone = errors.New("leader: SENZ_ZONE is required")
)

// New returns the Elector of the zone of conf. A clustered edge is
// elected through a Consul lock on its zone; a lone edge without
// SENZ_CONSUL_URI always leads.
func New(conf *settingpb.Config) (*Elector, error) {
	node, err := cluster.NodeId()
	if err != nil {
		return nil, err
	}

	uri := conf.Get(settingpb.Senz_SENZ_CONSUL_URI)
	if uri == "" {
		if conf.Clustered() {
			return nil, errNoConsul
		}

		szlog.Warnf("leader: no SENZ_CONSUL_URI, %s leads alone", node)
		return newElector(&standalone{node: node}), nil
	}

	// lowercased like the zone prefix of szrand ids, so SGN and sgn
	// campaign on the same key
	zone := strings.ToLower(conf.Get(settingpb.Senz_SENZ_ZONE))
	if zone == "" {
		return nil, errNoZone
	}

	client, err := consul.New(uri)
	if err != nil {
		return nil, fmt.Errorf("leader: consul: %w", err)
	}

	e := newElector(&consulBackend{
		client: client,
		key:    _keyPrefix + zone,
		node:   node,
	})
	e.zone = newZoneStore(client, zone)

	return e, nil
}

// consulBackend elects through a Consul lock: Consul serves it from its
// Raft log, so at most one session holds the key at a time.
type consulBackend struct {
	client *consul.Client
	key    string
	node   string
}

func (b *consulBackend) acquire(ctx context.Context) (*lease, error) {
	lock, err := b.client.Acquire(ctx, b.key, []byte(b.node), _sessionTTL)
	if err != nil {
		return nil, err
	}

	return &lease{
		term:    Term{Index: lock.Index(), Node: b.node},
		lost:    lock.Lost(),
		release: lock.Release,
	}, nil
}

// standalone leads at once and never loses: there is no one to share
// the zone with.
type standalone struct {
	node string
}

func (s *standalone) acquire(context.Context) (*lease, error) {
	return &lease{
		term:    Term{Index: 1, Node: s.node},
		release: func() error { return nil },
	}, nil
}
