// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routing

import (
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/szchains"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/shared/zlog"
)

func NewStandardRouter(store *memory.MemStore) szchains.ChainNode {
	return &StandardRouter{
		Node:  szchains.NewNode(),
		store: store,
	}
}

type StandardRouter struct {
	*szchains.Node
	store *memory.MemStore
}

func (r *StandardRouter) Handle(ctx szhttp.Context) error {
	target, err := r.store.Route().Match(ctx)
	if err != nil {
		zlog.Error("[edge] routing match error: ", err)
		return szhttp.NotFound(ctx)
	}

	rprx, ok := r.store.ReverseProxy().Load(target)
	if ok {
		rprx.Serve(ctx)
	}

	return nil
}
