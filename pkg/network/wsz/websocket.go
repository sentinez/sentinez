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
	corehttp "github.com/sentinez/core/http"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
	"github.com/sentinez/shared/sync"
)

func NewServer(appConf *settingpb.Config) *WebSocket {
	return &WebSocket{
		routers: sync.Map[string, func(corehttp.Context) error]{},
		meta:    appConf.GetMeta(),
	}
}

type WebSocket struct {
	routers sync.Map[string, func(corehttp.Context) error]
	meta    *typepb.XMeta
}

func (ws *WebSocket) HandlerFunc(
	path string, handler func(corehttp.Context) error) {

	_, ok := ws.routers.Load(path)
	if ok {
		panic("websocket: handler already exists for path: " + path)
	}

	ws.routers.Store(path, handler)
}

func (ws *WebSocket) ListenAndServe(
	addr string, options ...corehttp.ServerOption) error {

	ws.routers.Range(
		func(path string, handler func(corehttp.Context) error) bool {
			stdhttpx.HandlerFunc(path, handler)
			return true
		})

	ws.routers.Clear()

	return stdhttpx.ListenAndServe(addr, options...)
}

func (ws *WebSocket) Shutdown() error {
	return stdhttpx.Shutdown()
}
