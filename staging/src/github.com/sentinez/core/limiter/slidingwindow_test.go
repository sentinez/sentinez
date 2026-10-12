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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	_d     = 100 * time.Millisecond
	_size  = time.Second
	_limit = int64(10)
)

var (
	_t0  = time.Now().Truncate(_size)
	_t1  = _t0.Add(1 * _d)
	_t2  = _t0.Add(2 * _d)
	_t3  = _t0.Add(3 * _d)
	_t4  = _t0.Add(4 * _d)
	_t5  = _t0.Add(5 * _d)
	_t6  = _t0.Add(6 * _d)
	_t10 = _t0.Add(10 * _d)
	_t12 = _t0.Add(12 * _d)
	_t13 = _t0.Add(13 * _d)
	_t14 = _t0.Add(14 * _d)
	_t15 = _t0.Add(15 * _d)
	_t16 = _t0.Add(16 * _d)
	_t18 = _t0.Add(18 * _d)
	_t30 = _t0.Add(30 * _d)
)

type caseArg struct {
	t  time.Time
	n  int64
	ok bool
}

func TestLimiter_LocalWindow_SetLimit(t *testing.T) {
	lim := NewSlidingWindow(_size, _limit)

	got := lim.Limit()
	if got != _limit {
		t.Errorf("lim.Limit() = %d, want: %d", got, _limit)
	}

	newLimit := int64(12)
	lim.SetLimit(newLimit)
	got = lim.Limit()
	if got != newLimit {
		t.Errorf("lim.Limit() = %d, want: %d", got, newLimit)
	}
}

func TestLimiter_LocalWindow_AllowN(t *testing.T) {
	lim := NewSlidingWindow(_size, _limit)

	tests := []caseArg{
		// prev-window: empty, count: 0
		// curr-window: [t0, t0 + 1s), count: 0
		{_t0, 1, true},
		{_t1, 2, true},
		{_t2, 3, true},

		// count will be (1 + 2 + 3 + 5) = 11, so it fails
		{_t5, 5, false},

		// prev-window: [t0, t0 + 1s), count: 6
		// curr-window: [t10, t10 + 1s), count: 0
		{_t10, 2, true},

		// count will be (4/5*6 + 2 + 5) ≈ 11, so it fails
		{_t12, 5, false},

		{_t15, 5, true},

		// prev-window: [t30 - 1s, t30), count: 0
		// curr-window: [t30, t30 + 1s), count: 0
		{_t30, 10, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.ok, lim.AllowN(tt.t, tt.n),
				"lim.AllowN(%v, %v)", tt.t, tt.n)
		})
	}
}
