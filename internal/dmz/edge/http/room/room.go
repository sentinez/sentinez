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
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/sentinez/pkg/queue"
	"github.com/sentinez/shared/zlog"
)

var _ corechains.ChainNode = (*WaitingRoom)(nil)

func NewRoom(_ zlog.Level, _ *memory.MemStore) corechains.ChainNode {
	return &WaitingRoom{
		Node: corechains.NewNode(),
	}
}

type WaitingRoom struct {
	*corechains.Node
	_ *queue.Queue
}

func (wr *WaitingRoom) Handle(ctx corehttp.Context) error {

	return wr.HandleNext(ctx)
}
