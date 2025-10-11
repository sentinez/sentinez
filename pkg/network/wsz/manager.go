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
	"github.com/gorilla/websocket"
	"github.com/sentinez/sentinez/pkg/common/syncx"
	"github.com/sentinez/sentinez/pkg/zlog"
)

func NewManager() *Manager {
	return &Manager{
		clients: syncx.NewMap[string, *websocket.Conn](),
	}
}

type Manager struct {
	clients *syncx.Map[string, *websocket.Conn]
}

func (m *Manager) AddClient(id string, conn *websocket.Conn) {
	m.clients.Store(id, conn)

	zlog.Debugf("Manager.AddClient : client %s added", id)
}

func (m *Manager) RemoveClient(id string) {
	conn, ok := m.clients.Load(id)
	if !ok {
		zlog.Warnf("Manager.RemoveClient : client %s not found", id)
		return
	}

	if err := conn.Close(); err != nil {
		zlog.Errorf(
			"Manager.RemoveClient: error closing connection for client %s: %v",
			id, err)
	}

	m.clients.Delete(id)

	zlog.Debugf("Manager.RemoveClient : client %s removed", id)
}

func (m *Manager) Broadcast(message []byte) {
	m.clients.Range(func(id string, conn *websocket.Conn) bool {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			zlog.Errorf(
				"Manager.Broadcast: error sending message to client %s: %v",
				id, err)

			return false // stop iteration on error
		}

		zlog.Debugf("Manager.Broadcast: message sent to client %s", id)
		return true // continue iteration
	})
}
