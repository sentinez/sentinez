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

package requests

import (
	"context"

	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	corehttpreq "github.com/sentinez/sentinez/core/http/request"
	"github.com/sentinez/sentinez/shared/sync"
)

var (
	pool = sync.NewPool[corehttpreq.RequestContext]()
)

func New(ctx context.Context,
	req *edgepb.RequestContext) *corehttpreq.RequestContext {

	rctx := pool.Get()

	rctx.Req = req
	rctx.Ctx = ctx

	return rctx
}

func Free(rctx *corehttpreq.RequestContext) {
	if rctx == nil {
		return
	}

	rctx.Req = nil
	pool.Put(rctx)
}
