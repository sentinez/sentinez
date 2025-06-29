package main

import (
	"context"

	"github.com/sentinez/sentinez/cmd/discovery/v1/apps"
	"github.com/sentinez/sentinez/internal/core/discovery/v1"
	"github.com/sentinez/sentinez/pkg/core/sentinez/v1"
	"github.com/sentinez/sentinez/pkg/std/config"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func main() {
	if err := flags.Validate(apps.ParseFlag()); err != nil {
		zlog.Fatal(err)
	}

	app := sentinez.Build(discovery.New, config.Default, apps.ParseFlag)
	if err := app.Run(context.Background()); err != nil {
		zlog.Fatal(err)
	}
}
