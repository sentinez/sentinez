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

package ratelimiter

import (
	corechains "github.com/sentinez/core/chains"
	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/sentinez/internal/shared/mem/ratelimiter"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
	"github.com/sentinez/shared/zlog"
)

func NewLimiter(_ zlog.Level) corechains.ChainNode {
	return &Limiter{
		Node: corechains.NewNode(),
	}
}

type Limiter struct {
	*corechains.Node
}

func (l *Limiter) Handle(ctx corehttp.Context) error {
	limiter := ratelimiter.GetEngine().LoadContext(ctx)

	if !limiter.Allow(ctx.RequestIP()) {
		totalCount := limiter.Count(ctx.RequestIP()) + limiter.Limit()
		zlog.Debugf("limit exceeded %d - %s", totalCount, ctx.URI())
		return httpxcmn.TooManyRequests(ctx)
	}

	return l.HandleNext(ctx)
}
