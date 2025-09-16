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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/core/runner/v1/internal"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"google.golang.org/grpc/grpclog"

	"go.uber.org/fx"
)

var (
	logging zlog.Sugard

	options map[OptionType]any
)

type Engine interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func Main(start func(ctx context.Context) error, opts ...Option) {
	logging = zlog.NewDefaultConsole(zlog.LevelError)
	grpclog.SetLoggerV2(logging)

	options = make(map[OptionType]any)
	for _, opt := range opts {
		if opt == nil {
			continue
		}

		options[opt.Type()] = opt.Value()
	}

	ctn := container{}

	opt, ok := options[RunnerCtx]
	if !ok || opt == nil {
		zlog.Fatal("runner: missing runner context value")
	}

	rctx := opt.(*common.RunnerCtx)
	ctx := newContext(rctx)

	zlog.SetLogLevel(rctx.GetFlag().GetLogLevel())
	if rctx.GetFlag().GetEnvMode() != "dev" {
		internal.AppendOption(fx.NopLogger)
	}

	if err := start(ctx); err != nil {
		zlog.Fatal(err)
	}

	ctn.engine = fx.New(internal.Option())

	_ = ctn.Run(ctx)
}
