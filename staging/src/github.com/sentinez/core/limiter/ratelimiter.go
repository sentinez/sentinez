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
		records:    ssync.NewMap[string, *record](),
		windowSize: windowSize,
		limit:      limit,
		timeout:    timeout.Milliseconds(),
	}
}

type RateLimiter struct {
	records    *ssync.Map[string, *record]
	mu         sync.Mutex
	windowSize time.Duration
	limit      int64
	timeout    int64
}

// record is the per-key state, so that one key exceeding the limit does
// not block the others.
type record struct {
	window       *SlidingWindow
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

	rec, ok := r.records.Load(key)
	if !ok {
		rec = &record{window: NewSlidingWindow(r.windowSize, r.limit)}
		r.records.Store(key, rec)
	}

	nowMillis := now.UnixMilli()
	if rec.lastBlocking != 0 && rec.lastBlocking+r.timeout > nowMillis {
		rec.lastBlocking = nowMillis
		return false
	}

	if !rec.window.AllowN(now, n) {
		rec.lastBlocking = nowMillis
		return false
	}

	return true
}

func (r *RateLimiter) Count(key string) int64 {
	if r == nil {
		return 0
	}

	rec, ok := r.records.Load(key)
	if !ok {
		return 0
	}

	return rec.window.Count()
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
