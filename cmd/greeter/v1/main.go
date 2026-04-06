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

// Package main provides the entry point for the greeter service.
package main

import (
	"context"

	"github.com/sentinez/core/runner"
	"github.com/sentinez/sentinez"
	"github.com/sentinez/sentinez/cmd/greeter/v1/apps/config"
	"github.com/sentinez/sentinez/internal/core/greeter/v1"
)

func main() {
	conf := config.Config()
	grpc := greeter.NewService(conf.GetMeta())

	app := runner.NewApp(conf, sentinez.Code)
	app.Register(
		grpc.Start,
		grpc.Shutdown,
	)
	app.Run(context.Background())
}
