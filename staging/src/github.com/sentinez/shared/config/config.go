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
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

var (
	_envConf = make(map[string]string)
	_once    sync.Once
	_lock    sync.Mutex
)

// LoadEnv returns the environment.
func LoadEnv(
	envFile string, keys ...settingpb.Senz,
) (map[string]string, error) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("load env file: %w", err)
		}
	}

	_once.Do(func() {
		if len(keys) > 0 {
			for _, k := range keys {
				_envConf[k.String()] = os.Getenv(k.String())
			}

			return
		}

		for key := range settingpb.Senz_value {
			_envConf[key] = os.Getenv(key)
		}
	})

	return _envConf, nil
}

func GetEnv(key settingpb.Senz) string {
	_lock.Lock()
	defer _lock.Unlock()

	value, ok := _envConf[key.String()]
	if ok {
		return value
	}

	return ""
}

func SetEnv(key settingpb.Senz, value string) {
	_lock.Lock()
	defer _lock.Unlock()

	_envConf[key.String()] = value
}
