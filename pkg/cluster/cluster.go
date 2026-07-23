// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"sync"

	"github.com/hashicorp/memberlist"
	"github.com/sentinez/shared/zlog"
)

var (
	cluster *Cluster
	once    sync.Once
)

func Start() *Cluster {
	once.Do(func() {
		conf := memberlist.DefaultLANConfig()

		conf.Name = "node-1"
		conf.BindAddr = "0.0.0.0"
		conf.BindPort = 7946

		list, err := memberlist.Create(conf)
		if err != nil {
			zlog.Errorf("cluster: create memberlist err: %v", err)
			return
		}

		cluster = &Cluster{memlist: list}
	})

	return cluster
}

func Shutdown() error {
	return cluster.Shutdown()
}

type Cluster struct {
	memlist *memberlist.Memberlist
}

func (c *Cluster) Shutdown() error {
	if c == nil {
		return nil
	}

	return c.memlist.Shutdown()
}
