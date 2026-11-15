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

package centralsync

import (
	"context"
	"fmt"
	"io"

	"github.com/redis/go-redis/v9"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	"github.com/sentinez/sentinez/pkg/cluster"
)

// Subscriber opens the subscription to the Olric channel of centraldata;
// *cluster.Cluster is one.
type Subscriber interface {
	Subscribe(ctx context.Context, channel string) (*redis.PubSub, error)
}

var _ Subscriber = (*cluster.Cluster)(nil)

// Dialer connects to centraldata; the closer ends the connection. It is
// called for every load, so a centraldata that moved or was down is found.
type Dialer func(ctx context.Context,
) (centraldatapb.CentralDataServiceClient, io.Closer, error)

// NewSource returns the Source of centraldata, whose changes reach the edge
// through the Olric cluster.
func NewSource(cl Subscriber, dial Dialer) Source {
	return &source{cluster: cl, dial: dial}
}

type source struct {
	cluster Subscriber
	dial    Dialer
}

func (s *source) Subscribe(ctx context.Context,
) (<-chan *redis.Message, io.Closer, error) {
	sub, err := s.cluster.Subscribe(ctx, cluster.CentralDataChannel)
	if err != nil {
		return nil, nil, err
	}

	return sub.Channel(), sub, nil
}

// Sync streams all the synced tables.
func (s *source) Sync(ctx context.Context) (Stream, error) {
	client, conn, err := s.dial(ctx)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := client.Sync(ctx, &centraldatapb.SyncRequest{})
	if err != nil {
		cancel()
		_ = conn.Close()

		return nil, fmt.Errorf("open: %w", err)
	}

	return &endingStream{stream: stream, release: func() {
		cancel()
		_ = conn.Close()
	}}, nil
}

// endingStream releases the call and its connection once its snapshot, the
// last message, is read: the caller reads no further.
type endingStream struct {
	stream  centraldatapb.CentralDataService_SyncClient
	release func()
}

func (s *endingStream) Recv() (*centraldatapb.SyncResponse, error) {
	resp, err := s.stream.Recv()
	if err != nil || resp.GetDone() != nil {
		s.release()
	}

	return resp, err
}
