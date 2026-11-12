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
	"strings"

	centraldatasvc "github.com/sentinez/controlplane/apps/centraldata/v1/service"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsign"
	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/pkg/cluster"
)

// errNoSyncKey the edges would refuse what is published without a key, and
// a publisher without one would let anyone forge it.
var errNoSyncKey = errors.New("SENZ_SYNC_KEY is required with " +
	"SENZ_OLRIC_ADDRESSES")

// NewPublisher returns the Publisher of the changes of centraldata to the
// Olric cluster of the edges (SENZ_OLRIC_ADDRESSES), signed with
// SENZ_SYNC_KEY. Without addresses it publishes nothing, which is for
// running centraldata alone.
func NewPublisher(conf *settingpb.Config) (centraldatasvc.Publisher, error) {
	addrs := splitList(conf.Get(settingpb.Senz_SENZ_OLRIC_ADDRESSES))
	if len(addrs) == 0 {
		szlog.Warnf("centraldata: no SENZ_OLRIC_ADDRESSES, " +
			"the changes are not published")

		return nopPublisher{}, nil
	}

	key, err := syncKey(conf)
	if err != nil {
		return nil, err
	}

	pub, err := cluster.NewPublisher(addrs)
	if err != nil {
		return nil, err
	}

	return &signedPublisher{pub: pub, key: key}, nil
}

func syncKey(conf *settingpb.Config) ([]byte, error) {
	encoded := conf.Get(settingpb.Senz_SENZ_SYNC_KEY)
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

	return p.pub.Publish(ctx, cluster.CentralDataChannel, data)
}

// nopPublisher publishes nothing.
type nopPublisher struct{}

func (nopPublisher) Publish(context.Context, *centraldatapb.Change) error {
	return nil
}

// splitList the non-empty items of the comma-separated list.
func splitList(list string) []string {
	var items []string
	for _, item := range strings.Split(list, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}

	return items
}
