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

	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
	"github.com/sentinez/sentinez/pkg/runner/internal"
	"github.com/sentinez/sentinez/pkg/zlog"
	"google.golang.org/grpc/grpclog"

	"go.uber.org/fx"
)

type Engine interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func Main(appConf *configspb.AppConfig, start func(ctx context.Context) error) {
	logging := zlog.NewDefaultConsole(zlog.LevelError)

	grpclog.SetLoggerV2(logging)
	zlog.SetLogLevel(appConf.GetFlag().GetLogLevel())

	if appConf.GetFlag().GetEnvMode() != "dev" {
		internal.AppendOption(fx.NopLogger)
	}

	ctx := newContext(appConf)
	if err := start(ctx); err != nil {
		zlog.Fatal(err)
	}

	ctn := container{engine: fx.New(internal.Option())}
	_ = ctn.Run(ctx)
}
