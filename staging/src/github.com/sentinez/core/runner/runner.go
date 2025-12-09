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

package runner

import (
	"context"

	"github.com/sentinez/core/runner/internal"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/shared/zlog"

	"go.uber.org/fx"
)

type Engine interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func Serve(ctx context.Context, app *App) {
	if app.start != nil {
		if err := app.start(app.conf); err != nil {
			zlog.Fatal(err)
		}
	}

	app.Inject(func() *confpb.Config {
		return app.conf
	})

	ctn := container{engine: fx.New(internal.Option())}
	_ = ctn.Run(ctx)
}
