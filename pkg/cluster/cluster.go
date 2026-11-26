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
	"os"
	"sync"

	"github.com/hashicorp/go-sockaddr"
	"github.com/olric-data/olric"
	"github.com/olric-data/olric/config"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szrand"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

const _nodeIdLen = 6

var (
	_dict   *Cluster
	_once   sync.Once
	_newErr error
)

// newConfig is the Olric config of appConf; started is called once the
// node accepts connections, clustered or not. privateIP picks the IP of an
// unspecified bind.
func newConfig(appConf *settingpb.Config, started func(),
	privateIP privateIPFunc) (*config.Config, error) {
	conf := config.New("local")
	conf.Started = started

	var err error
	conf.BindAddr, conf.BindPort, err = olricBind(appConf, privateIP)
	if err != nil {
		return nil, err
	}

	if appConf.Clustered() {
		conf.Peers = []string{
			appConf.Get(settingpb.Sz_SZ_MEMBERSHIP_ADDRESS),
		}

		ml := conf.MemberlistConfig
		if ml.BindAddr, ml.BindPort, err = discoveryBind(appConf,
			privateIP); err != nil {
			return nil, err
		}
	}

	if err := secure(conf, appConf); err != nil {
		return nil, err
	}

	return conf, nil
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

	conf, err := newConfig(appConf, c.markReady, sockaddr.GetPrivateIP)
	if err != nil {
		return nil, err
	}

	if err := c.open(conf); err != nil {
		return nil, err
	}

	return c, nil
}

type Cluster struct {
	db     *olric.Olric
	client *olric.EmbeddedClient
	// addr is the Olric address of this node, its name in the cluster
	addr string
	// node names this process in the cluster
	node      string
	ready     chan struct{}
	readyOnce sync.Once
	done      chan struct{}
}

// open creates the Olric node of conf.
func (c *Cluster) open(conf *config.Config) error {
	db, err := olric.New(conf)
	if err != nil {
		return fmt.Errorf("cluster: new olric: %w", err)
	}

	// olric.New names the node after its bind address.
	c.db, c.client = db, db.NewEmbeddedClient()
	c.addr = conf.MemberlistConfig.Name

	return nil
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
