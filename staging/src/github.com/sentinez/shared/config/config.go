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
	"os"
	"sync"

	"github.com/joho/godotenv"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	"github.com/sentinez/shared/zlog"
)

var (
	envConf = make(map[string]string)
	once    sync.Once
	lock    sync.Mutex
)

// LoadEnv returns the environment.
func LoadEnv(envFile string, keys ...settingpb.Senz) map[string]string {
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			zlog.Fatalf("error loading environment file: err=%v", err)
		}
	}

	once.Do(func() {
		if len(keys) > 0 {
			for _, k := range keys {
				envConf[k.String()] = os.Getenv(k.String())
			}

			return
		}

		for key := range settingpb.Senz_value {
			envConf[key] = os.Getenv(key)
		}
	})

	return envConf
}

func GetEnv(key settingpb.Senz) string {
	lock.Lock()
	defer lock.Unlock()

	value, ok := envConf[key.String()]
	if ok {
		return value
	}

	return ""
}

func SetEnv(key settingpb.Senz, value string) {
	lock.Lock()
	defer lock.Unlock()

	envConf[key.String()] = value
}
