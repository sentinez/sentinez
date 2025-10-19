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

// Package realtime provides the realtime server for the apiserver
package realtime

import (
	"context"

	realtimehdl "github.com/sentinez/sentinez/internal/realtime/handlers"
	"github.com/sentinez/sentinez/pkg/network/wsz"
	"github.com/sentinez/sentinez/pkg/runner/v1"
)

func New(ws *wsz.WebSocket) *Realtime {
	return &Realtime{
		core: ws,
	}
}

type Realtime struct {
	core *wsz.WebSocket
}

func (r *Realtime) router() {
	r.core.HandlerFunc("/ws", realtimehdl.Handler)
}

// Start implements runner.Server.
func (r *Realtime) Start(ctx context.Context) error {
	appConf := runner.GetAppConfig(ctx)
	// register the route with websocket handler
	r.router()
	return r.core.ListenAndServe(appConf.GetEnvConf().GetHttpAddress())
}

// Shutdown implements runner.Server.
func (r *Realtime) Shutdown(_ context.Context) error {
	return r.core.Shutdown()
}
