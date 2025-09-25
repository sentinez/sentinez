// Copyright 2025 Sentinez Labs.
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

	"github.com/sentinez/sentinez/pkg/std/zlog"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy ProxyConfig `yaml:"proxy"`
}

type ProxyConfig struct {
	Namespace string        `yaml:"namespace"`
	Routes    []RouteConfig `yaml:"routes"`
}

type RouteConfig struct {
	MatchPrefix string `yaml:"match_prefix"`
	Target      string `yaml:"target"`
	Rewrite     string `yaml:"rewrite,omitempty"`
}

func LoadRouteConfig(filename string) *Config {
	data, err := os.ReadFile(filename)
	if err != nil {
		zlog.Fatal(err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		zlog.Fatal(err)
	}

	return &cfg
}
