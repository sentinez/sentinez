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

	greeterpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"

	"github.com/sentinez/sentinez/cmd/greeter/v1/apps"
	"github.com/sentinez/sentinez/internal/core/greeter/v1"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/config"
)

// Build and run main application with environment variable
// Remember to inject all layers of the application by
// runner.Inject() function
//
// Example:
//
// _ = runner.Inject(controllers.New)
func main() {
	flag := apps.ParseFlag()
	conf := config.Load(flag.GetEnvFile())

	app := runner.New(greeter.NewService).Build(
		func(service *greeter.Service) (runner.Server, error) {
			service.SetPref(greeterpb.GetMetaGreeter(), conf, flag)
			return greeter.New(service), nil
		},
	)

	_ = app.Run(context.Background())
}
