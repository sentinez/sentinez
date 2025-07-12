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

	wshandlers "github.com/sentinez/sentinez/internal/websocket/handlers"
	"github.com/sentinez/sentinez/pkg/core/runner/v1"
	"github.com/sentinez/sentinez/pkg/core/websocket"
	"github.com/sentinez/sentinez/pkg/std/names"
	"github.com/sentinez/sentinez/pkg/std/version"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var _ runner.Server = (*WebSocket)(nil)

func New(ws *websocket.WebSocket) runner.Server {
	return &WebSocket{
		core: ws,
	}
}

type WebSocket struct {
	core *websocket.WebSocket
}

func (w *WebSocket) router() {
	w.core.HandlerFunc("/ws", wshandlers.Handler)
}

// Start implements runner.Server.
func (w *WebSocket) Start(_ context.Context) error {
	// register the route with websocket handler
	w.router()

	version.ASCII("SENTINEZ // WEB SOCKET", names.WebSocket.String())
	zlog.Infof("[HTTP] starting server %s", ":7778")

	return w.core.ListenAndServe(":7778")
}

// Shutdown implements runner.Server.
func (w *WebSocket) Shutdown(_ context.Context) error {
	return w.core.Shutdown()
}
