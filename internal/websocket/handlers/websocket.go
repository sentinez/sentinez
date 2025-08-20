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

// Package wshandlers provides the handlers for the websocket server
package wshandlers

import (
	"fmt"

	"github.com/sentinez/sentinez/internal/websocket/manager"
	httpx1 "github.com/sentinez/sentinez/pkg/core/net/httpx/h1"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func Handler(ctx httpx1.Context) error {
	conn, err := ctx.Upgrade()
	if err != nil {
		zlog.Errorf("wshandlers.Handler upgrade error: %v", err)
		return err
	}
	defer func() { _ = conn.Close() }()

	clientID := ctx.Request().URL.Query().Get("id")
	if clientID == "" {
		zlog.Error("wshandlers.Handler missing client ID")
		return errors.ErrInvalidData
	}

	manager.Manager().AddClient(clientID, conn)
	defer manager.Manager().RemoveClient(clientID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			zlog.Errorf("wshandlers.Handler read error from %s: %v",
				clientID, err)
			break
		}

		zlog.Debugf("wshandlers.Handler received from %s: %s", clientID, msg)
		manager.Manager().
			Broadcast(fmt.Appendf(nil, "From %s: %s", clientID, msg))
	}

	return nil
}
