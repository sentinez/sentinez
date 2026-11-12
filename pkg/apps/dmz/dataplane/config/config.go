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
	dataplanepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/dataplane/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/pkg/appconfig"
	"github.com/sentinez/sentinez/pkg/apps/dmz/dataplane/flags"
)

var _provider = appconfig.New(appconfig.Spec{
	Meta:  dataplanepb.GetMetaDataplane(),
	Flags: flags.Parse,
	Keys: []settingpb.Senz{
		settingpb.Senz_SENZ_HOSTNAME,
		settingpb.Senz_SENZ_ADDRESS,
		settingpb.Senz_SENZ_SECRET_KEY,
		settingpb.Senz_SENZ_OTLP_INTERNAL_ENDPOINT,
	},
})

// Config loads the config once and returns it: the flags, and the env from
// the environment and from Consul (see appconfig).
func Config() (*settingpb.Config, error) {
	return _provider.Config()
}
