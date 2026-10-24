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

package room

import (
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/shared/szqueue"
)

// WaitingRoomId is the chain stage name of WaitingRoom.
const WaitingRoomId = "ROM"

var _ chains.Plugin = (*WaitingRoom)(nil)

type WaitingRoom struct {
	_ *szqueue.Queue
}

func (*WaitingRoom) ID() string { return WaitingRoomId }

func (*WaitingRoom) Handle(ctx szhttp.Context, next chains.Handler) error {
	return next.Handle(ctx)
}
