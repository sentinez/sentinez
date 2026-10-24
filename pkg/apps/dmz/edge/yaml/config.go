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
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/szlog"
	"google.golang.org/protobuf/encoding/protojson"
)

var errMissingSetting = errors.New("missing 'setting' in configuration")

type Config struct {
	Setting *edgepb.Setting `yaml:"setting"`
}

func LoadSetting(appConf *settingpb.Config) (*edgepb.Setting, error) {
	data, err := os.ReadFile(appConf.GetFlag().GetProxyConfig())
	if err != nil {
		return nil, fmt.Errorf("read proxy config: %w", err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse proxy config: %w", err)
	}

	settingRaw, ok := raw["setting"]
	if !ok {
		return nil, errMissingSetting
	}

	szlog.Debugf("setting: %v", settingRaw)

	settingJSON, err := json.Marshal(settingRaw)
	if err != nil {
		return nil, fmt.Errorf("encode setting: %w", err)
	}

	setting := &edgepb.Setting{}
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}

	if err := unmarshaler.Unmarshal(settingJSON, setting); err != nil {
		return nil, fmt.Errorf("decode setting: %w", err)
	}

	return setting, nil
}
