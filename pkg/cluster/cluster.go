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
	"os"
	"strconv"
	"sync"

	"github.com/olric-data/olric"
	"github.com/olric-data/olric/config"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szrand"
)

const _nodeIdLen = 6

var (
	_dict   *Cluster
	_once   sync.Once
	_newErr error
)

// newConfig is the Olric config of appConf; started is called once the
// node accepts connections, clustered or not.
func newConfig(appConf *settingpb.Config, started func()) *config.Config {
	membership := appConf.Get(settingpb.Senz_SENZ_MEMBERSHIP_ADDRESS)
	discovery := appConf.Get(settingpb.Senz_SENZ_DISCOVERY_ADDRESS)

	conf := config.New("local")
	conf.Started = started

	if appConf.Clustered() {
		conf.Peers = []string{membership}

		host, port, err := net.SplitHostPort(discovery)
		if err == nil {
			nport, _ := strconv.Atoi(port)

			conf.MemberlistConfig.BindAddr = host
			conf.MemberlistConfig.BindPort = nport
			return conf
		}

		szlog.Warnf("cluster: discovery address invalid, use default")
		conf.MemberlistConfig.BindAddr = defaults.DiscoveryAddress
		conf.MemberlistConfig.BindPort = defaults.DiscoveryPort
	}

	return conf
}

func NewCluster(appConf *settingpb.Config) (*Cluster, error) {
	_once.Do(func() { _dict, _newErr = newCluster(appConf) })

	return _dict, _newErr
}

func newCluster(appConf *settingpb.Config) (*Cluster, error) {
	node, err := NodeId()
	if err != nil {
		return nil, err
	}

	c := &Cluster{
		node:  node,
		ready: make(chan struct{}),
		done:  make(chan struct{}),
	}

	db, err := olric.New(newConfig(appConf, c.markReady))
	if err != nil {
		return nil, fmt.Errorf("cluster: new olric: %w", err)
	}

	c.db = db
	c.client = db.NewEmbeddedClient()

	return c, nil
}

type Cluster struct {
	db     *olric.Olric
	client *olric.EmbeddedClient
	// node names this process in the cluster
	node      string
	ready     chan struct{}
	readyOnce sync.Once
	done      chan struct{}
}

// Ready is closed once the node accepts connections.
func (c *Cluster) Ready() <-chan struct{} { return c.ready }

// Node names this process in the cluster.
func (c *Cluster) Node() string { return c.node }

// Start serves the cluster in the background until Shutdown is called.
func (c *Cluster) Start() {
	go func() {
		defer close(c.done)

		if err := c.db.Start(); err != nil {
			szlog.Errorf("cluster: start: %v", err)
		}
	}()
}

// markReady closes Ready, once: Olric may report started again.
func (c *Cluster) markReady() {
	c.readyOnce.Do(func() {
		close(c.ready)
		szlog.Infof("cluster:olric: ready to accept connection")
	})
}

// NodeId returns a new name for this process: the hostname alone is not
// unique when several edges share a host.
func NodeId() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("cluster: hostname: %w", err)
	}

	suffix, err := szrand.RandomString(_nodeIdLen)
	if err != nil {
		return "", fmt.Errorf("cluster: node id: %w", err)
	}

	return host + "-" + suffix, nil
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
