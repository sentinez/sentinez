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

	"github.com/sentinez/sentinez/cmd/websocket/apps/config"
	"github.com/sentinez/sentinez/internal/websocket"
	wscore "github.com/sentinez/sentinez/pkg/core/net/wsz"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
)

func main() {
	runner.Main(config.Config(), func(ctx context.Context) error {
		conf := runner.GetAppConfig(ctx)
		wsSrv := wscore.NewServer(conf.GetMeta())
		ws := websocket.New(wsSrv)

		runner.OnStart(ws.Start)
		runner.OnStop(ws.Shutdown)

		return nil
	})
}
