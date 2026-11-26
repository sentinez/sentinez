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

package cluster

import (
	"encoding/base64"
	"fmt"
	"net"

	"github.com/olric-data/olric/config"
	"github.com/sentinez/pkg/szlog"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// secure sets the password of the Olric port (SZ_OLRIC_PASSWORD), which
// Olric also gives its own client to the other nodes, and the key of the
// memberlist gossip (SZ_GOSSIP_KEY). It warns of what is left open on a
// port others reach.
func secure(conf *config.Config, appConf *settingpb.Config) error {
	conf.Authentication = &config.Authentication{
		Password: appConf.Get(settingpb.Sz_SZ_OLRIC_PASSWORD),
	}

	if conf.Authentication.Password == "" && !loopback(conf.BindAddr) {
		szlog.Warnf("cluster: no SZ_OLRIC_PASSWORD: whoever reaches "+
			"%s:%d reads and writes the dmaps", conf.BindAddr, conf.BindPort)
	}

	if !appConf.Clustered() {
		return nil
	}

	key, err := gossipKey(appConf)
	if err != nil {
		return err
	}

	if key == nil {
		szlog.Warnf("cluster: no SZ_GOSSIP_KEY: the gossip is in clear " +
			"and anyone may join the cluster")
	}

	conf.MemberlistConfig.SecretKey = key

	return nil
}

// gossipKey the memberlist key of SZ_GOSSIP_KEY, nil when unset.
func gossipKey(appConf *settingpb.Config) ([]byte, error) {
	encoded := appConf.Get(settingpb.Sz_SZ_GOSSIP_KEY)
	if encoded == "" {
		return nil, nil
	}

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("cluster: SZ_GOSSIP_KEY: %w", err)
	}

	switch len(key) {
	case 16, 24, 32:
		return key, nil
	default:
		return nil, fmt.Errorf("cluster: SZ_GOSSIP_KEY: %d bytes, "+
			"want 16, 24 or 32", len(key))
	}
}

// loopback reports whether host is a loopback IP.
func loopback(host string) bool {
	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}
