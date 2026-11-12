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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/sentinez/api/client/consul"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/defaults"
	"google.golang.org/protobuf/encoding/protojson"
)

// _consulTimeout bounds the read of the proxy config from Consul, which
// runs before the edge starts.
const _consulTimeout = 5 * time.Second

var errMissingSetting = errors.New("missing 'setting' in configuration")

type Config struct {
	Setting *edgepb.Setting `yaml:"setting"`
}

// rawGetter reads a value of the Consul KV.
type rawGetter interface {
	GetRaw(ctx context.Context, key string) ([]byte, error)
}

// LoadSetting reads the proxy config. With SENZ_CONSUL_URI it is the one at
// the Consul key defaults.ProxyConfigKey, the very one the apiserver
// imports; without the URI, or when Consul has no such key, it is the file
// named by the proxy_config flag.
func LoadSetting(appConf *settingpb.Config) (*edgepb.Setting, error) {
	path := appConf.GetFlag().GetProxyConfig()

	var data []byte
	var err error

	if uri := appConf.Get(settingpb.Senz_SENZ_CONSUL_URI); uri != "" {
		data, err = readConsulProxy(uri, path)
	} else {
		data, err = os.ReadFile(path)
	}

	if err != nil {
		return nil, fmt.Errorf("read proxy config: %w", err)
	}

	return Parse(data)
}

func readConsulProxy(uri, path string) ([]byte, error) {
	client, err := consul.New(uri)
	if err != nil {
		return nil, fmt.Errorf("consul: %w", err)
	}

	return readProxy(client, path)
}

// readProxy reads the proxy config from Consul, or from the file at path
// when Consul has no such key.
func readProxy(getter rawGetter, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), _consulTimeout)
	defer cancel()

	data, err := getter.GetRaw(ctx, defaults.ProxyConfigKey)
	if errors.Is(err, consul.ErrKeyNotFound) {
		szlog.Warnf("proxy config: no consul key %q, reading %s",
			defaults.ProxyConfigKey, path)

		return os.ReadFile(path)
	}
	if err != nil {
		return nil, fmt.Errorf("consul key %q: %w",
			defaults.ProxyConfigKey, err)
	}

	return data, nil
}

// Parse decodes the 'setting' block of a proxy config (proxy.yaml).
func Parse(data []byte) (*edgepb.Setting, error) {
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
