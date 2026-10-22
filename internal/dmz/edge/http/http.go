// Copyright 2025 Sentinéz Labs.
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

package http

import (
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/funcs"
	"github.com/sentinez/sentinez/internal/memory"
)

// NewHandler links the edge chain nodes reading ms and returns its head;
// opts are passed to every node (see option.WithLogger,
// option.WithHostname).
func NewHandler(
	ms *memory.MemStore, opts ...option.Option) chains.ChainNode {

	handlers := funcs.New(ms, opts...)
	if len(handlers) == 0 {
		return nil
	}

	// first node is head of all chains, link the rest in order
	income := handlers[0]
	curr := income
	for _, handler := range handlers[1:] {
		curr = curr.SetNext(handler)
	}

	return income
}
