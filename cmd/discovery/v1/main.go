package main

import (
	"context"

	"github.com/sentinez/sentinez/cmd/discovery/v1/apps"
	"github.com/sentinez/sentinez/internal/core/discovery/v1"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/std/config"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func main() {
	if err := flags.Validate(apps.ParseFlag()); err != nil {
		zlog.Fatal(err)
	}

	app := runner.New(discovery.NewService).
		Build(func(service *discovery.Service) (runner.Server, error) {
			return discovery.
				New(service, config.Default(), apps.ParseFlag()), nil
		})

	if err := app.Run(context.Background()); err != nil {
		zlog.Fatal(err)
	}
}
