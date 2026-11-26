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

package centraldata

import (
	"context"
	"errors"

	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsign"
	"github.com/sentinez/pkg/sztls"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/pkg/cluster"
	"github.com/sentinez/sentinez/pkg/cluster/channel"
)

// errNoSyncKey the edges would refuse what is published without a key, and
// a publisher without one would let anyone forge it.
var errNoSyncKey = errors.New("SZ_SYNC_KEY is required with SZ_OLRIC_URI")

// NewPublisher returns the Publisher of the changes of centraldata to the
// Olric cluster of the edges (SZ_OLRIC_URI), signed with SZ_SYNC_KEY.
// Without a URI it publishes nothing, which is for running centraldata
// alone.
func NewPublisher(conf *settingpb.Config) (centraldatasvc.Publisher, error) {
	uri := conf.Get(settingpb.Sz_SZ_OLRIC_URI)
	if uri == "" {
		szlog.Warnf("centraldata: no SZ_OLRIC_URI, " +
			"the changes are not published")

		return nopPublisher{}, nil
	}

	key, err := syncKey(conf)
	if err != nil {
		return nil, err
	}

	// olrics:// reaches the TLS proxies in front of the nodes
	ep, err := sztls.ParseURI(uri, "olric", "olrics", sztls.FilesOf(conf))
	if err != nil {
		return nil, err
	}

	pub, err := cluster.NewPublisher(ep.Hosts, cluster.WithTLS(ep.TLS),
		cluster.WithPassword(conf.Get(settingpb.Sz_SZ_OLRIC_PASSWORD)))
	if err != nil {
		return nil, err
	}

	return &signedPublisher{pub: pub, key: key}, nil
}

func syncKey(conf *settingpb.Config) ([]byte, error) {
	encoded := conf.Get(settingpb.Sz_SZ_SYNC_KEY)
	if encoded == "" {
		return nil, errNoSyncKey
	}

	return szsign.ParseKey(encoded)
}

// signedPublisher publishes the changes in an envelope signed with key.
type signedPublisher struct {
	pub *cluster.Publisher
	key []byte
}

func (p *signedPublisher) Publish(ctx context.Context,
	change *centraldatapb.Change) error {

	data, err := szsign.Seal(p.key, change)
	if err != nil {
		return err
	}

	return p.pub.Publish(ctx, channel.CentralData, data)
}

// nopPublisher publishes nothing.
type nopPublisher struct{}

func (nopPublisher) Publish(context.Context, *centraldatapb.Change) error {
	return nil
}
