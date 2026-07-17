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

	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	"github.com/sentinez/shared/zlog"
)

var envConf *settingpb.EnvConfig
var once sync.Once

func Env() *settingpb.EnvConfig {
	return envConf
}

func SetEnv(env *settingpb.EnvConfig) {
	once.Do(func() {
		envConf = env
	})
}

// LoadEnv returns the environment.
func LoadEnv(envFile string) *settingpb.EnvConfig {
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			zlog.Fatalf("error loading environment file: err=%v", err)
		}
	}

	once.Do(func() {
		envConf = &settingpb.EnvConfig{
			TimescaleUri:   os.Getenv("SENZ_TIMESCALE_URI"),
			PostgresUri:    os.Getenv("SENZ_POSTGRES_URI"),
			ClickhouseUri:  os.Getenv("SENZ_CLICKHOUSE_URI"),
			ConsulUri:      os.Getenv("SENZ_CONSUL_URI"),
			SecretKey:      os.Getenv("SENZ_SECRET_KEY"),
			GatewayAddress: os.Getenv("SENZ_GATEWAY_ADDRESS"),
			Hostname:       os.Getenv("SENZ_HOSTNAME"),
			HttpAddress:    os.Getenv("SENZ_HTTP_ADDRESS"),
			ClientOrigin:   os.Getenv("SENZ_CLIENT_ORIGIN"),
			GrpcAddress:    os.Getenv("SENZ_GRPC_ADDRESS"),
		}
	})

	return envConf
}
