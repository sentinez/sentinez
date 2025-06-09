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

	"github.com/sentinez/sentinez/cmd/greeter/v1/apps"
	"github.com/sentinez/sentinez/internal/core/greeter/v1"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/std/config"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

// Build and run main application with environment variable
// Remember to inject all layers of the application by
// sentinez.Inject() function
//
// Example:
//
// _ = sentinez.Inject(controllers.New)
func main() {
	if err := flags.Validate(apps.ParseFlag()); err != nil {
		zlog.Fatal(err)
	}

	app := sentinez.Build(greeter.New, config.Default, apps.ParseFlag)
	if err := app.Run(context.Background()); err != nil {
		zlog.Fatal(err)
	}
}
