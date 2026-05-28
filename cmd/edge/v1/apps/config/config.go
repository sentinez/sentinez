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
	"sync"

	"github.com/sentinez/shared/config"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	edgeflags "github.com/sentinez/sentinez/cmd/edge/v1/apps/flags"
)

var (
	once    sync.Once
	appConf *confpb.Config
)

func Config() *confpb.Config {
	once.Do(func() {
		flag := edgeflags.Parse()
		envConf := config.LoadEnv(flag.GetEnvFile())
		appConf = &confpb.Config{
			Meta: edgepb.GetMetaEdge(),
			Env:  envConf,
			Flag: flag,
		}
	})

	return appConf
}
