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

	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	realtimehdl "github.com/sentinez/sentinez/pkg/dmz/realtime/handlers"
	"github.com/sentinez/sentinez/pkg/network/wsz"
)

func New(conf *confpb.Config, ws *wsz.WebSocket) *Realtime {
	return &Realtime{
		core: ws,
		conf: conf,
	}
}

type Realtime struct {
	core *wsz.WebSocket
	conf *confpb.Config
}

func (r *Realtime) router() {
	r.core.HandlerFunc("/ws", realtimehdl.Handler)
}

func (r *Realtime) Start() error {
	// register the route with websocket handler
	r.router()
	return r.core.ListenAndServe(r.conf.GetEnv().GetHttpAddress())
}

func (r *Realtime) Shutdown(_ context.Context) error {
	return r.core.Shutdown()
}
