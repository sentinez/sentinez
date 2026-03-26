// Copyright 2025 Sentinéz Labs.
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

package edgeyaml

import (
	"os"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	"github.com/sentinez/shared/zlog"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Setting *edgepb.Setting `yaml:"setting"`
}

func LoadSetting(appConf *confpb.Config) *edgepb.Setting {
	data, err := os.ReadFile(appConf.GetFlag().GetProxyConfig())
	if err != nil {
		zlog.Fatal(err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		zlog.Fatal(err)
	}

	return cfg.Setting
}
