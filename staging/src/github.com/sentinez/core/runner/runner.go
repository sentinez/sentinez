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

	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	"github.com/sentinez/shared/zlog"
	"go.uber.org/fx"
	"google.golang.org/grpc/grpclog"
)

func NewApp[T any](appConf *settingpb.Config, scopeName string) *App[T] {
	logging := zlog.NewConsole(scopeName, zlog.LevelError)
	grpclog.SetLoggerV2(logging)

	level := zlog.ToLevel(appConf.GetFlag().GetLogLevel())
	zlog.SetScopeLogLevel(scopeName, level)
	ctx := NewContext[T](appConf)

	return &App[T]{
		ctx: ctx,
	}
}

type App[T any] struct {
	ctx *Context[T]
}

func (a *App[T]) Main(main func(*Context[T])) {
	main(a.ctx)

	ctn := container{engine: fx.New(a.ctx.opts...)}
	if err := ctn.Run(context.Background()); err != nil {
		zlog.Fatal(err)
	}
}
