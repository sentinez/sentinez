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

package corelimiter

import (
	"sync"
	"time"

	ssync "github.com/sentinez/shared/sync"
)

func NewRateLimiter(timeout time.Duration,
	windowSize time.Duration, limit int64) *RateLimiter {
	return &RateLimiter{
		records:    ssync.NewMap[string, *SlidingWindow](),
		windowSize: windowSize,
		limit:      limit,
		timeout:    timeout.Milliseconds(),
	}
}

type RateLimiter struct {
	records      *ssync.Map[string, *SlidingWindow]
	mu           sync.Mutex
	windowSize   time.Duration
	limit        int64
	timeout      int64
	lastBlocking int64
}

func (r *RateLimiter) Allow(key string) bool {
	if r == nil {
		return true
	}

	return r.AllowN(key, time.Now(), 1)
}

func (r *RateLimiter) AllowN(key string, now time.Time, n int64) bool {
	if r == nil {
		return true
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	limiter, ok := r.records.Load(key)
	if !ok {
		limiter = NewSlidingWindow(r.windowSize, r.limit)
		r.records.Store(key, limiter)
	}

	if r.lastBlocking != 0 && r.lastBlocking+r.timeout > now.UnixMilli() {
		r.lastBlocking = now.UnixMilli()
		return false
	}

	if !limiter.AllowN(now, n) {
		r.lastBlocking = now.UnixMilli()
		return false
	}

	return r.lastBlocking+r.timeout <= now.UnixMilli()
}

func (r *RateLimiter) Count(key string) int64 {
	if r == nil {
		return 0
	}

	limiter, ok := r.records.Load(key)
	if !ok {
		return 0
	}

	return limiter.Count()
}

func (r *RateLimiter) Size() time.Duration {
	if r == nil {
		return 0
	}

	return r.windowSize
}

func (r *RateLimiter) Limit() int64 {
	if r == nil {
		return 0
	}

	return r.limit
}
