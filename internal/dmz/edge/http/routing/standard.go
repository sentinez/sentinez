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
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/shared/zlog"
)

func NewStandardRouter(store *memory.MemStore) corechains.ChainNode {
	return &StandardRouter{
		Node:  corechains.NewNode(),
		store: store,
	}
}

type StandardRouter struct {
	*corechains.Node
	store *memory.MemStore
}

func (r *StandardRouter) Handle(ctx corehttp.Context) error {
	target, err := r.store.Route().Match(ctx)
	if err != nil {
		zlog.Error("[edge] routing match error: ", err)
		return corehttp.NotFound(ctx)
	}

	rprx, ok := r.store.ReverseProxy().Load(target)
	if ok {
		rprx.Serve(ctx)
	}

	return nil
}
