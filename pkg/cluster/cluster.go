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
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/hashicorp/memberlist"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/zlog"
)

var (
	cluster *Cluster
	once    sync.Once
)

func New(meta *typepb.XMeta, address string) *Cluster {
	once.Do(func() {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			zlog.Errorf("cluster: address invalid %s", address)
		}
		bindPort, _ := strconv.Atoi(port)

		name := fmt.Sprintf("%s-%s", meta.GetServiceKey(), address)

		conf := memberlist.DefaultLANConfig()
		conf.Name = name
		conf.BindAddr = host
		conf.BindPort = bindPort

		list, err := memberlist.Create(conf)
		if err != nil {
			zlog.Errorf("cluster: create memberlist err: %v", err)
			return
		}

		cluster = &Cluster{
			memlist: list,
			Name:    name,
			Addr:    host,
			Port:    bindPort,
		}
	})

	return cluster
}

type Cluster struct {
	memlist *memberlist.Memberlist
	Name    string
	Addr    string
	Port    int
}

func (c *Cluster) Shutdown() error {
	if c == nil {
		return nil
	}

	return c.memlist.Shutdown()
}
