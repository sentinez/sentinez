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
	dict *Dictionary
	once sync.Once
)

func NewDictionary(confpb *settingpb.Config) *Dictionary {
	once.Do(func() {
		conf := config.New("local")

		db, err := olric.New(conf)
		if err != nil {
			zlog.Fatal(err)
		}

		membership := confpb.GetDefault(
			settingpb.Senz_SENZ_MEMBERSHIP_ADDRESS,
			defaults.MembershipAddress,
		)

		dict = &Dictionary{
			db:      db,
			cluster: cluster.New(edgepb.GetMetaEdge(), membership),
		}
	})

	return dict
}

type Dictionary struct {
	cluster *cluster.Cluster
	db      *olric.Olric
}

func (d *Dictionary) Start() {
	go func() {
		if err := d.db.Start(); err != nil {
			zlog.Fatal(err)
		}
	}()
}

func (d *Dictionary) Shutdown(ctx context.Context) {
	_ = d.db.Shutdown(ctx)
	_ = d.cluster.Shutdown()
}
