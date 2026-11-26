// Copyright 2025 Duc-Hung Ho.
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

package config

import (
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/pkg/appconfig"
	edgeflags "github.com/sentinez/sentinez/pkg/apps/dmz/edge/flags"
)

var _provider = appconfig.New(appconfig.Spec{
	Meta:  edgepb.GetMetaEdge(),
	Flags: edgeflags.Parse,
	Keys: []settingpb.Sz{
		settingpb.Sz_SZ_ADDRESS,
		settingpb.Sz_SZ_SECRET_KEY,
		settingpb.Sz_SZ_MEMBERSHIP_ADDRESS,
		settingpb.Sz_SZ_DISCOVERY_ADDRESS,
		settingpb.Sz_SZ_OLRIC_ADDRESS,
		settingpb.Sz_SZ_OLRIC_PASSWORD,
		settingpb.Sz_SZ_GOSSIP_KEY,
		settingpb.Sz_SZ_OTLP_INTERNAL_URI,
		settingpb.Sz_SZ_OTLP_EXTERNAL_URI,
		settingpb.Sz_SZ_KAFKA_URI,
		settingpb.Sz_SZ_CONSUL_URI,
		settingpb.Sz_SZ_ZONE,
		// centraldata: the edge follows it when it holds a service token
		settingpb.Sz_SZ_SERVICE_TOKEN,
		settingpb.Sz_SZ_SYNC_KEY,
		settingpb.Sz_SZ_TLS_CA,
		settingpb.Sz_SZ_TLS_CERT,
		settingpb.Sz_SZ_TLS_KEY,
		// opens the certificates centraldata gives the edge to serve
		settingpb.Sz_SZ_DATA_KEY,
	},
})

// Config loads the config once and returns it: the flags, and the env from
// the environment and from Consul (see appconfig).
func Config() (*settingpb.Config, error) {
	return _provider.Config()
}
