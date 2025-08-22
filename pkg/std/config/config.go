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

// Package config provides the configs for the service.
package config

import (
	"os"
	"sync"

	_ "github.com/joho/godotenv/autoload" // load .env file automatically
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
)

var conf *common.Config
var once sync.Once

// Default returns the environment.
func Default() *common.Config {
	once.Do(func() {
		conf = &common.Config{
			TimescaleUri:  getENV(common.SNTZENV_SNTZENV_TIMESCALEDB),
			PostgresUri:   getENV(common.SNTZENV_SNTZENV_POSTGRES),
			ClickhouseUri: getENV(common.SNTZENV_SNTZENV_CLICKHOUSE),
			SecretKey:     getENV(common.SNTZENV_SNTZENV_SECRET_KEY),
		}
	})

	return conf
}

func getENV(key common.SNTZENV) string {
	return os.Getenv(key.String())
}
