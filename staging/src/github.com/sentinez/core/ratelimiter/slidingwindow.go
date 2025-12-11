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

// Here is a summary of the requirements for the system:
// - Accurately limit excessive requests.
// - Low latency. The rate limiter should not slow down HTTP response time.
// - Use as little memory as possible.
// - Distributed rate limiting. The rate limiter can be shared across
//   multiple servers or processes.
// - Exception handling. Show clear exceptions to users when their requests
//   are throttled.
// - High fault tolerance. If there are any problems with the rate limiter
//   (for example, a cache server goes offline), it does not affect the
//   entire system.
//
// We implement base on sliding window counter algorithm
// Pros
// - It smooths out spikes in traffic because the rate is based on the average
//   rate of the previous window.
// - Memory efficient.
//
// Cons
// - It only works for not-so-strict look back window. It is an approximation
//   of the actual rate because it assumes requests in the previous window
//   are evenly distributed. However, this problem may not be as bad as
//	 it seems. According to experiments done by Cloudflare [10], only 0.003%
//   of requests are wrongly allowed or rate limited among 400 million requests
//
// ref:
// https://bytebytego.com/courses/system-design-interview/design-a-rate-limiter

package ratelimiter

import (
	"sync"
	"time"
)

type Limiter interface {
	Size() time.Duration
	Limit() int64
	SetLimit(limit int64)

	Count() int64

	Allow() bool
	AllowN(now time.Time, n int64) bool
}

var _ Limiter = (*SlidingWindow)(nil)

func NewSlidingWindow(size time.Duration, limit int64) *SlidingWindow {
	return &SlidingWindow{
		size:  size,
		limit: limit,
		curr:  NewLocalWindow(),
		prev:  NewLocalWindow(),
	}
}

type SlidingWindow struct {
	size  time.Duration
	limit int64
	mu    sync.Mutex

	curr Window
	prev Window
}

func (sw *SlidingWindow) Count() int64 {
	return sw.curr.Count()
}

func (sw *SlidingWindow) Size() time.Duration {
	return sw.size
}

func (sw *SlidingWindow) Limit() int64 {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.limit
}

func (sw *SlidingWindow) SetLimit(limit int64) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.limit = limit
}

func (sw *SlidingWindow) Allow() bool {
	return sw.AllowN(time.Now(), 1)
}

func (sw *SlidingWindow) AllowN(now time.Time, n int64) bool {
	if sw == nil {
		return true
	}

	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.advance(now)

	elapsed := now.Sub(sw.curr.Start())
	weight := float64(sw.size-elapsed) / float64(sw.size)
	count := int64(weight*float64(sw.prev.Count())) + sw.curr.Count()

	// Trigger the possible sync behaviour.
	defer sw.curr.Sync(now)

	if count+n > sw.limit {
		return false
	}

	sw.curr.AddCount(n)
	return true
}

// advance updates the current/previous
// windows resulting from the passage of time.
func (sw *SlidingWindow) advance(now time.Time) {
	// Calculate the start boundary of the expected current-window.
	newCurrStart := now.Truncate(sw.size)

	diffSize := newCurrStart.Sub(sw.curr.Start()) / sw.size
	if diffSize >= 1 {
		// The current-window is at least one-window-size
		// behind the expected one.

		newPrevCount := int64(0)
		if diffSize == 1 {
			// The new previous-window will overlap with the old current-window,
			// so it inherits the count.
			//
			// Note that the count here may be not accurate, since it is only a
			// SNAPSHOT of the current-window's count, which in itself tends to
			// be inaccurate due to the asynchronous
			// nature of the sync behaviour.
			newPrevCount = sw.curr.Count()
		}
		sw.prev.Reset(newCurrStart.Add(-sw.size), newPrevCount)

		// The new current-window always has zero count.
		sw.curr.Reset(newCurrStart, 0)
	}
}
