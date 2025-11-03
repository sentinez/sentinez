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
	"github.com/sentinez/sentinez/core/networks"
	"github.com/sentinez/sentinez/pkg/common/syncx"
)

var (
	pool = syncx.NewPool[networks.RequestContext]()
)

func New(ctx context.Context,
	req *edgepb.RequestContext) *networks.RequestContext {

	rctx := pool.Get()

	rctx.Req = req
	rctx.Ctx = ctx

	return rctx
}

func Free(rctx *networks.RequestContext) {
	if rctx == nil {
		return
	}

	rctx.Req = nil
	pool.Put(rctx)
}
