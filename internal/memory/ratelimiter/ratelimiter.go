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
	"sync"

	corehttp "github.com/sentinez/core/http"
	corelimiter "github.com/sentinez/core/limiter"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	once        sync.Once
	limiterInst *Limiter
	mu          sync.Mutex
)

func New() *Limiter {
	once.Do(func() {
		limiterInst = &Limiter{
			space: ssync.NewMap[string, *corelimiter.RateLimiter](),
		}
	})

	return limiterInst
}

func GetEngine() *Limiter {
	return limiterInst
}

type Limiter struct {
	space *ssync.Map[string, *corelimiter.RateLimiter]
}

func (lim *Limiter) Store(serverName string, l *corelimiter.RateLimiter) {
	lim.space.Store(serverName, l)
}

func (lim *Limiter) Load(serverName string) *corelimiter.RateLimiter {
	expr, ok := lim.space.Load(serverName)
	if !ok {
		return nil
	}

	return expr
}

func (lim *Limiter) LoadContext(ctx corehttp.Context) *corelimiter.RateLimiter {
	if lim == nil {
		return nil
	}

	hCtx, ok := corehttp.GetRequestContext(ctx)
	if !ok {
		return nil
	}

	zlog.Debugf("[edge] hit limiter cached %s", hCtx.GetServerName())
	return lim.Load(hCtx.GetServerName())
}

func Store(namespace string, l *corelimiter.RateLimiter) {
	mu.Lock()
	defer mu.Unlock()

	if limiterInst == nil {
		limiterInst = New()
	}

	limiterInst.Store(namespace, l)
}
