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

// Package flags provides the app setting for apiserver service
package flags

import (
	"sync"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/apiserver/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/std/stdflag"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/spf13/pflag"
)

var onceAPIServer sync.Once

// Parse flag args for apiserver service
func Parse() *common.Flag {
	onceAPIServer.Do(func() {
		stdflag.Get().ApiSpecsPath = "./resources/api/specs/v1"
		stdflag.Get().SwaggerPath = "./resources/api/swagger"
		stdflag.Get().EnvFile = "./cmd/apiserver/.env"

		pflag.StringVar(&stdflag.Get().ApiSpecsPath, "api-specs-path",
			stdflag.Get().GetApiSpecsPath(), "openapi specification path")

		pflag.StringVar(&stdflag.Get().SwaggerPath, "swagger-path",
			stdflag.Get().GetSwaggerPath(), "swagger user interface path")

		pflag.StringVar(&stdflag.Get().EnvFile, "env-file",
			stdflag.Get().GetEnvFile(), "environment variable config file")

		stdflag.Parse(apiserver.GetMetaApiserver())
	})

	if err := stdflag.Validate(stdflag.Get()); err != nil {
		zlog.Fatal(err)
	}

	return stdflag.Get()
}
