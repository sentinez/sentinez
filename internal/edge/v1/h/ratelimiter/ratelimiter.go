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
	"time"

	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/core/ratelimiter"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
	"github.com/sentinez/shared/zlog"
)

func New() chains.Handler {
	return &Limiter{
		BaseHandler: chains.New(),
		limiter:     ratelimiter.NewSlidingWindow(time.Minute, 10),
	}
}

type Limiter struct {
	*chains.BaseHandler
	limiter ratelimiter.Limiter
}

func (l *Limiter) Handle(ctx corehttp.Context) error {
	if !l.limiter.Allow() {
		zlog.Debugf("limit exceeded %d - %s", l.limiter.Count(), ctx.URI())
		return httpxcmn.TooManyRequests(ctx)
	}

	return l.HandleNext(ctx)
}
