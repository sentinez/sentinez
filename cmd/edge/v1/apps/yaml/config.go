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

type Routes struct {
	Routes []Route `yaml:"routes"`
}

type Route struct {
	Location  string `yaml:"location"`
	ProxyPass string `yaml:"proxy_pass"`
	Static    bool   `yaml:"static"`
}

func LoadRoutesFromYAML(filename string) *Routes {
	data, err := os.ReadFile(filename)
	if err != nil {
		zlog.Fatal(err)
	}

	var cfg Routes
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		zlog.Fatal(err)
	}

	return &cfg
}
