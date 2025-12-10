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
	"testing"
	"time"
)

const (
	d     = 100 * time.Millisecond
	size  = time.Second
	limit = int64(10)
)

var (
	t0  = time.Now().Truncate(size)
	t1  = t0.Add(1 * d)
	t2  = t0.Add(2 * d)
	t3  = t0.Add(3 * d)
	t4  = t0.Add(4 * d)
	t5  = t0.Add(5 * d)
	t6  = t0.Add(6 * d)
	t10 = t0.Add(10 * d)
	t12 = t0.Add(12 * d)
	t13 = t0.Add(13 * d)
	t14 = t0.Add(14 * d)
	t15 = t0.Add(15 * d)
	t16 = t0.Add(16 * d)
	t18 = t0.Add(18 * d)
	t30 = t0.Add(30 * d)
)

type caseArg struct {
	t  time.Time
	n  int64
	ok bool
}

func TestLimiter_LocalWindow_SetLimit(t *testing.T) {
	lim := NewSlidingWindow(size, limit)

	got := lim.Limit()
	if got != limit {
		t.Errorf("lim.Limit() = %d, want: %d", got, limit)
	}

	newLimit := int64(12)
	lim.SetLimit(newLimit)
	got = lim.Limit()
	if got != newLimit {
		t.Errorf("lim.Limit() = %d, want: %d", got, newLimit)
	}
}

func TestLimiter_LocalWindow_AllowN(t *testing.T) {
	lim := NewSlidingWindow(size, limit)

	cases := []caseArg{
		// prev-window: empty, count: 0
		// curr-window: [t0, t0 + 1s), count: 0
		{t0, 1, true},
		{t1, 2, true},
		{t2, 3, true},

		// count will be (1 + 2 + 3 + 5) = 11, so it fails
		{t5, 5, false},

		// prev-window: [t0, t0 + 1s), count: 6
		// curr-window: [t10, t10 + 1s), count: 0
		{t10, 2, true},

		// count will be (4/5*6 + 2 + 5) ≈ 11, so it fails
		{t12, 5, false},

		{t15, 5, true},

		// prev-window: [t30 - 1s, t30), count: 0
		// curr-window: [t30, t30 + 1s), count: 0
		{t30, 10, true},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			ok := lim.AllowN(c.t, c.n)
			if ok != c.ok {
				t.Errorf("lim.AllowN(%v, %v) = %v, want: %v",
					c.t, c.n, ok, c.ok)
			}
		})
	}
}
