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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	"github.com/sentinez/sentinez/pkg/client/names"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/version"
	"github.com/spf13/pflag"
)

var onceAPIServer sync.Once

// apiServerFlags global variable
var apiServerFlags = &common.FlagAPIServer{
	ApiSpecsPath: "resources/api/specs/v1",
	SwaggerPath:  "resources/api/swagger",
	Address:      ":9000",
}

// ParseFlag flag args for apiserver service
func ParseFlag() *common.FlagAPIServer {
	onceAPIServer.Do(func() {
		console := version.FigureGen(
			"SENTINEZ // API SERVER", names.APIServer.String())

		flags.SetConsole(console, names.APIServer, "dev")

		pflag.StringVarP(&apiServerFlags.Address, "address", "a",
			apiServerFlags.GetAddress(), "host address")

		pflag.StringVar(&apiServerFlags.ApiSpecsPath, "api-specs",
			apiServerFlags.GetApiSpecsPath(), "openapi specification path")

		pflag.StringVar(&apiServerFlags.SwaggerPath, "swagger-ui",
			apiServerFlags.GetSwaggerPath(), "swagger ui path")
	})

	_ = flags.Parse()

	return apiServerFlags
}
