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

package distrib

import (
	"context"
	"sync"

	"github.com/olric-data/olric"
	"github.com/olric-data/olric/config"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/sentinez/pkg/cluster"
	"github.com/sentinez/shared/zlog"
)

var (
	dict      *Cluster
	once      sync.Once
	dictReady = make(chan bool, 1)
)

func StartCluster(conf *settingpb.Config) {
	_ = NewCluster(conf)
	dict.Start()
}

func ShutdownCluster(ctx context.Context) {
	dict.Shutdown(ctx)
}

func NewCluster(appConf *settingpb.Config) *Cluster {
	once.Do(func() {
		conf := config.New("local")
		conf.Started = func() {
			dictReady <- true
			zlog.Infof("cluster:olric: ready to accept connection")
		}

		db, err := olric.New(conf)
		if err != nil {
			zlog.Fatal(err)
		}

		membership := appConf.GetDefault(
			settingpb.Senz_SENZ_MEMBERSHIP_ADDRESS,
			defaults.MembershipAddress,
		)

		dict = &Cluster{
			db:      db,
			cluster: cluster.New(edgepb.GetMetaEdge(), membership),
			client:  db.NewEmbeddedClient(),
		}
	})

	return dict
}

type Cluster struct {
	cluster *cluster.Cluster
	db      *olric.Olric
	client  *olric.EmbeddedClient
}

func (d *Cluster) Start() {
	go func() {
		if err := d.db.Start(); err != nil {
			zlog.Fatal(err)
		}
	}()
}

func (d *Cluster) Shutdown(ctx context.Context) {
	_ = d.db.Shutdown(ctx)
	_ = d.cluster.Shutdown()
}
