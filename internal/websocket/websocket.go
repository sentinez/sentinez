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

// Package websocket provides the websocket server for the apiserver
package websocket

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	wshandlers "github.com/sentinez/sentinez/internal/websocket/handlers"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/core/wsz"
)

var _ runner.Server = (*WebSocket)(nil)

func New(ws *wsz.WebSocket, flags *common.FlagWS) runner.Server {
	return &WebSocket{
		core: ws,
		flag: flags,
	}
}

type WebSocket struct {
	core *wsz.WebSocket
	flag *common.FlagWS
}

func (w *WebSocket) router() {
	w.core.HandlerFunc("/ws", wshandlers.Handler)
}

// Start implements runner.Server.
func (w *WebSocket) Start(_ context.Context) error {
	// register the route with websocket handler
	w.router()
	return w.core.ListenAndServe(w.flag.GetAddress())
}

// Shutdown implements runner.Server.
func (w *WebSocket) Shutdown(_ context.Context) error {
	return w.core.Shutdown()
}
