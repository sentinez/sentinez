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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	"github.com/sentinez/sentinez/pkg/zlog"
)

var envConf *common.EnvConfig
var once sync.Once

// LoadEnv returns the environment.
func LoadEnv(envFile string) *common.EnvConfig {
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			zlog.Fatalf("error loading environment file: err=%v", err)
		}
	}

	once.Do(func() {
		envConf = &common.EnvConfig{
			TimescaleUri:  os.Getenv("SNTZ_TIMESCALE_URI"),
			PostgresUri:   os.Getenv("SNTZ_POSTGRES_URI"),
			ClickhouseUri: os.Getenv("SNTZ_CLICKHOUSE_URI"),
			ConsulUri:     os.Getenv("SNTZ_CONSUL_URI"),
			SecretKey:     os.Getenv("SNTZ_SECRET_KEY"),
			GatewayAddr:   os.Getenv("SNTZ_GATEWAY_ADDR"),
			Hostname:      os.Getenv("SNTZ_HOSTNAME"),
			Address:       os.Getenv("SNTZ_ADDRESS"),
			PasskeyOrigin: os.Getenv("SNTZ_PASSKEY_ORIGIN"),
		}
	})

	return envConf
}
