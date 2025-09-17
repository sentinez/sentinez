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

package wsz

import (
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/common/sync"
	httpx1 "github.com/sentinez/sentinez/pkg/core/net/httpx/h1"
	"github.com/sentinez/sentinez/pkg/std/version"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func NewServer(meta *common.SntzMeta) *WebSocket {
	return &WebSocket{
		routers: sync.Map[string, func(httpx1.Context) error]{},
		meta:    meta,
	}
}

type WebSocket struct {
	routers sync.Map[string, func(httpx1.Context) error]
	meta    *common.SntzMeta
}

func (ws *WebSocket) HandlerFunc(
	path string, handler func(httpx1.Context) error) {

	_, ok := ws.routers.Load(path)
	if ok {
		panic("websocket: handler already exists for path: " + path)
	}

	ws.routers.Store(path, handler)
}

func (ws *WebSocket) ListenAndServe(addr string) error {
	ws.routers.Range(
		func(path string, handler func(httpx1.Context) error) bool {
			httpx1.HandlerFunc(path, handler)
			return true
		})

	ws.routers.Clear()

	version.INFO(
		ws.meta.GetServiceName(),
		ws.meta.GetServiceKey(),
	)

	zlog.Infof("%s >>> running on %s",
		color.Blue.Add("ws"),
		color.Magenta.Add(addr),
	)

	return httpx1.ListenAndServe(addr)
}

func (ws *WebSocket) Shutdown() error {
	return httpx1.Shutdown()
}
