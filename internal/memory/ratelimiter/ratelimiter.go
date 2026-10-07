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
	corerule "github.com/sentinez/core/rules"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	once        sync.Once
	limiterInst *Limiter
)

func New() *Limiter {
	once.Do(func() {
		limiterInst = &Limiter{
			space: ssync.NewMap[string, []Entry](),
		}
	})

	return limiterInst
}

// Entry is a rate limiter applied to the requests matching Eval.
type Entry struct {
	Eval    corerule.EvalFunc
	Rule    *secrulepb.SecRule
	Limiter *corelimiter.RateLimiter
}

type Limiter struct {
	space *ssync.Map[string, []Entry]
}

// Store replaces the limiters of serverName.
func (lim *Limiter) Store(serverName string, entries []Entry) {
	if len(entries) == 0 {
		lim.space.Delete(serverName)
		return
	}

	lim.space.Store(serverName, entries)
}

// Load returns the limiters of serverName. The result is shared and must
// not be modified.
func (lim *Limiter) Load(serverName string) []Entry {
	entries, _ := lim.space.Load(serverName)
	return entries
}

func (lim *Limiter) LoadContext(ctx corehttp.Context) []Entry {
	if lim == nil {
		return nil
	}

	zlog.Debugf("edge: hit limiter cached %s", ctx.X().GetNamespace())
	return lim.Load(ctx.X().GetNamespace())
}
