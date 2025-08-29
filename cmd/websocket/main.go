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

package main

import (
	"context"

	wspb "github.com/sentinez/sentinez/api/gen/go/sentinez/ws/v1"
	"github.com/sentinez/sentinez/cmd/websocket/apps"
	"github.com/sentinez/sentinez/internal/websocket"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	wscore "github.com/sentinez/sentinez/pkg/core/wsz"
	"github.com/sentinez/sentinez/pkg/std/config"
	"github.com/sentinez/sentinez/pkg/std/flags"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func main() {
	flag := apps.ParseFlag()
	if err := flags.Validate(flag); err != nil {
		zlog.Fatal(err)
	}

	conf := config.Load(flag.GetEnvFile())

	app := runner.New(wscore.New).
		Build(func(ws *wscore.WebSocket) (runner.Server, error) {
			ws.Metadata = wspb.GetMetaWs()
			return websocket.New(ws, flag, conf), nil
		})

	_ = app.Run(context.Background())
}
