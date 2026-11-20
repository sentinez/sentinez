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

package edge

import (
	"context"
	"errors"
	"io"

	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsign"
	"github.com/sentinez/sentinez/api/client"
	"github.com/sentinez/sentinez/api/client/connection"
	"github.com/sentinez/sentinez/api/client/options"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/centralsync"
	"github.com/sentinez/sentinez/pkg/cluster/leader"
)

// errCentralConfig the edge holds a service token but cannot use it.
var errCentralConfig = errors.New("centraldata sync is not configured")

// Lead campaigns for the leadership of the zone until ctx ends. The leader
// keeps the zone Status in Consul and, when the edge holds a service token
// (SZ_SERVICE_TOKEN), follows centraldata for the whole cluster: it
// writes the settings of the tables to the synced map every edge loads.
func (s *RevProxyServer) Lead(ctx context.Context) error {
	if !s.beginLead() {
		return nil
	}
	defer s.leading.Done()

	var follow leader.Task
	if s.conf.Get(settingpb.Sz_SZ_SERVICE_TOKEN) != "" {
		follow = s.followCentralData
	}

	return s.elector.Run(ctx, leader.All(s.elector.PublishStatus(), follow))
}

// beginLead counts a running Lead, unless the server is shutting down.
func (s *RevProxyServer) beginLead() bool {
	s.leadMu.Lock()
	defer s.leadMu.Unlock()

	if s.leadClosed {
		return false
	}

	s.leading.Add(1)

	return true
}

// endLead stops new Leads and waits for the running one, so nothing
// writes the cluster once it stops.
func (s *RevProxyServer) endLead() {
	s.leadMu.Lock()
	s.leadClosed = true
	s.leadMu.Unlock()

	s.leading.Wait()
}

// followCentralData is the leader.Task that loads the tables of centraldata
// and keeps them current, until the term ends. It never ends the term: a
// configuration error does not mend by campaigning again, and the settings
// already in the cluster serve meanwhile.
func (s *RevProxyServer) followCentralData(ctx context.Context,
	term leader.Term) error {
	// the map is written once the edge synced it
	select {
	case <-s.synced:
	case <-ctx.Done():
		return nil
	}

	follower, err := s.newFollower()
	if err != nil {
		szlog.Errorf("edge: centraldata: %v", err)

		return nil
	}
	defer follower.Close()

	szlog.Infof("edge: centraldata: followed by %s, term %d",
		term.Node, term.Index)

	// Start returns once the tables are loaded, or when the term ends; the
	// follower logs its retries meanwhile
	if err := follower.Start(ctx); err == nil {
		<-ctx.Done()
	}

	return nil
}

func (s *RevProxyServer) newFollower() (*centralsync.Follower, error) {
	key, err := szsign.ParseKey(s.conf.Get(settingpb.Sz_SZ_SYNC_KEY))
	if err != nil {
		return nil, errors.Join(errCentralConfig, err)
	}

	cluster := s.memstore.Cluster()
	if cluster == nil {
		return nil, errors.Join(errCentralConfig,
			errors.New("no cluster"))
	}

	source := centralsync.NewSource(cluster, s.dialCentralData)

	return centralsync.New(source, s.memstore.CentralSink(), key, 0), nil
}

// dialCentralData finds centraldata in consul and connects to it with the
// internal mTLS and the service token of the edge.
func (s *RevProxyServer) dialCentralData(context.Context,
) (centraldatapb.CentralDataServiceClient, io.Closer, error) {
	tls := connection.TLSFromConfig(s.conf)
	opt := &options.Options{
		ConsulURL: s.conf.Get(settingpb.Sz_SZ_CONSUL_URI),
	}

	cl, conn, err := client.NewCentralData(opt, tls,
		s.conf.Get(settingpb.Sz_SZ_SERVICE_TOKEN))
	if err != nil {
		return nil, nil, err
	}

	return cl, conn, nil
}
