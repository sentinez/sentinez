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

// Package main provides the entry point for the Sentinez.
package main

import (
	"context"

	"github.com/sentinez/sentinez/cmd/apiserver/apps/config"
	"github.com/sentinez/sentinez/internal/apiserver"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/sentinez/sentinez/pkg/runner"
)

// This is the sentinez apiserver application, it will automatically
// connect to other services via gRPC. Run the application along with
// other services in the cmd/ directory.The application provides APIs
// for users through a single HTTP gateway following the REST API standard.
// The application uses gRPC to connect to other services.Additionally,
// the system provides a Swagger UI interface for users to easily interact
// with the system through a web interface.
//
// Run the application using the Makefile command
//
//	make apiserver.run // start sentinez apiserver
//	make <service>.run // start service
func main() {
	runner.Main(config.Config(), func(ctx context.Context) error {
		conf := runner.GetAppConfig(ctx)
		httpSrv := httpx.NewServer(conf.GetMeta())
		server := apiserver.New(httpSrv)

		runner.OnStart(server.Start)
		runner.OnStop(server.Shutdown)

		return nil
	})
}
