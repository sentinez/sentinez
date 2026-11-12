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
	"time"

	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsign"
	"github.com/sentinez/sentinez/api/client"
	"github.com/sentinez/sentinez/api/client/connection"
	"github.com/sentinez/sentinez/api/client/options"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/centralsync"
)

// _centralTimeout bounds the wait for the first load of the tables of
// centraldata: past it the edge serves with what it has, and the load goes
// on in the background.
const _centralTimeout = 10 * time.Second

// errCentralConfig the edge holds a service token but cannot use it.
var errCentralConfig = errors.New("centraldata sync is not configured")

// followCentralData loads the tables of centraldata into the edge and keeps
// them current, when the edge holds a service token (SENZ_SERVICE_TOKEN).
// It never fails the edge: the file and the settings of the cluster serve
// meanwhile.
func (s *Server) followCentralData() {
	if s.conf.Get(settingpb.Senz_SENZ_SERVICE_TOKEN) == "" {
		return
	}

	follower, err := s.newFollower()
	if err != nil {
		szlog.Errorf("edge: centraldata: %v", err)

		return
	}

	s.follower = follower

	ctx, cancel := context.WithTimeout(context.Background(),
		_centralTimeout)
	defer cancel()

	if err := follower.Start(ctx); err != nil {
		szlog.Warnf("edge: centraldata: %v, serving without it for now", err)
	}
}

func (s *Server) newFollower() (*centralsync.Follower, error) {
	key, err := szsign.ParseKey(s.conf.Get(settingpb.Senz_SENZ_SYNC_KEY))
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
func (s *Server) dialCentralData(context.Context,
) (centraldatapb.CentralDataServiceClient, io.Closer, error) {
	tls := connection.TLSFromConfig(s.conf)
	opt := &options.Options{
		ConsulURL: s.conf.Get(settingpb.Senz_SENZ_CONSUL_URI),
	}

	cl, conn, err := client.NewCentralData(opt, tls,
		s.conf.Get(settingpb.Senz_SENZ_SERVICE_TOKEN))
	if err != nil {
		return nil, nil, err
	}

	return cl, conn, nil
}
