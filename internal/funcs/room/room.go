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
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/sentinez/pkg/queue"
)

var _ chains.ChainNode = (*WaitingRoom)(nil)

func NewRoom(
	_ *memory.MemStore, _ ...option.Option) chains.ChainNode {

	return &WaitingRoom{
		Node: chains.NewNode(),
	}
}

type WaitingRoom struct {
	*chains.Node
	_ *queue.Queue
}

func (wr *WaitingRoom) Handle(ctx szhttp.Context) error {

	return wr.HandleNext(ctx)
}
