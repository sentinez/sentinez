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
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/olric-data/olric"
	"github.com/olric-data/olric/config"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/shared/zlog"
)

var (
	_dict      *Cluster
	_once      sync.Once
	_dictReady = make(chan bool, 1)
	_newErr    error
)

func newConfig(appConf *settingpb.Config) *config.Config {
	membership := appConf.Get(settingpb.Senz_SENZ_MEMBERSHIP_ADDRESS)
	discovery := appConf.Get(settingpb.Senz_SENZ_DISCOVERY_ADDRESS)

	conf := config.New("local")
	if len(membership) != 0 && len(discovery) != 0 {
		conf.Peers = []string{membership}

		host, port, err := net.SplitHostPort(discovery)
		if err == nil {
			nport, _ := strconv.Atoi(port)

			conf.MemberlistConfig.BindAddr = host
			conf.MemberlistConfig.BindPort = nport
			return conf
		}

		zlog.Warnf("cluster: discovery address invalid, use default")
		conf.MemberlistConfig.BindAddr = defaults.DiscoveryAddress
		conf.MemberlistConfig.BindPort = defaults.DiscoveryPort
	}

	conf.Started = func() {
		_dictReady <- true
		zlog.Infof("cluster:olric: ready to accept connection")
	}

	return conf
}

func NewCluster(appConf *settingpb.Config) (*Cluster, error) {
	_once.Do(func() {
		db, err := olric.New(newConfig(appConf))
		if err != nil {
			_newErr = fmt.Errorf("cluster: new olric: %w", err)
			return
		}

		_dict = &Cluster{
			db:     db,
			client: db.NewEmbeddedClient(),
			done:   make(chan struct{}),
		}
	})

	return _dict, _newErr
}

type Cluster struct {
	db     *olric.Olric
	client *olric.EmbeddedClient
	done   chan struct{}
}

func (c *Cluster) NewDMap[T any](name string) *DMap[T] {
	m := &DMap[T]{
		name: name,
		temp: make(chan objKV[T], _defaultChanBufferSize),
	}
	go m.wait()
	return m
}

// Start serves the cluster in the background until Shutdown is called.
func (c *Cluster) Start() {
	go func() {
		defer close(c.done)

		if err := c.db.Start(); err != nil {
			zlog.Errorf("cluster: start: %v", err)
		}
	}()
}

// Shutdown stops the cluster and waits for Start to return or ctx to end.
func (c *Cluster) Shutdown(ctx context.Context) error {
	err := c.db.Shutdown(ctx)
	if err != nil {
		return err
	}

	select {
	case <-c.done:
	case <-ctx.Done():
	}

	return nil
}
