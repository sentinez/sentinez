// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package apps provides the app setting for apiserver service
package apps

import (
	"sync"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/apiserver/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/spf13/pflag"
)

var onceAPIServer sync.Once

// ParseFlag flag args for apiserver service
func ParseFlag() *common.Flag {
	onceAPIServer.Do(func() {
		flags.Get().ApiSpecsPath = "resources/api/specs/v1"
		flags.Get().SwaggerPath = "resources/api/swagger"
		flags.Get().EnvFile = "./cmd/apiserver/.env"

		pflag.StringVar(&flags.Get().ApiSpecsPath, "api-specs-path",
			flags.Get().GetApiSpecsPath(), "openapi specification path")

		pflag.StringVar(&flags.Get().SwaggerPath, "swagger-path",
			flags.Get().GetSwaggerPath(), "swagger user interface path")

		pflag.StringVar(&flags.Get().EnvFile, "env-file",
			flags.Get().GetEnvFile(), "environment variable config file")

		flags.Parse(apiserver.GetMetaApiserver())
	})

	return flags.Get()
}
