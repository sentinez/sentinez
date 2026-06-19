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

// Package realtimehdl provides the handlers for the realtime server
package realtimehdl

import (
	"fmt"

	corehttp "github.com/sentinez/core/http"
	realtimemnt "github.com/sentinez/sentinez/pkg/acz/realtime/manager"
	"github.com/sentinez/shared/errorx"
	"github.com/sentinez/shared/zlog"
)

func Handler(ctx corehttp.Context) error {
	conn, err := ctx.Upgrade()
	if err != nil {
		zlog.Errorf("wshandlers.Handler upgrade error: %v", err)
		return err
	}
	defer func() { _ = conn.Close() }()

	zlog.Debugf("wshandlers.Handler full path %s", ctx.Path())

	clientID := ctx.Query("id")
	if clientID == "" {
		zlog.Error("wshandlers.Handler missing client ID")
		return errorx.ErrInvalidData
	}

	realtimemnt.Manager().AddClient(clientID, conn)
	defer realtimemnt.Manager().RemoveClient(clientID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			zlog.Errorf("wshandlers.Handler read error from %s: %v",
				clientID, err)
			break
		}

		zlog.Debugf("wshandlers.Handler received from %s: %s", clientID, msg)
		realtimemnt.Manager().
			Broadcast(fmt.Appendf(nil, "From %s: %s", clientID, msg))
	}

	return nil
}
