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
)

func TestRateLimiter_BlockIsPerKey(t *testing.T) {
	lim := NewRateLimiter(time.Second, size, 2)

	for _, ok := range []bool{true, true, false} {
		if got := lim.AllowN("1.1.1.1", t1, 1); got != ok {
			t.Fatalf("AllowN(1.1.1.1) = %v, want: %v", got, ok)
		}
	}

	if !lim.AllowN("2.2.2.2", t2, 1) {
		t.Errorf("AllowN(2.2.2.2) = false, want: true")
	}
}

func TestRateLimiter_BlockUntilTimeout(t *testing.T) {
	lim := NewRateLimiter(500*time.Millisecond, size, 1)

	cases := []caseArg{
		{t1, 1, true},
		// exceeds the limit, blocked for 500ms from t2
		{t2, 1, false},
		// still within the timeout of the last rejected request
		{t5, 1, false},
		// timeout elapsed and the window has moved on
		{t15, 1, true},
	}

	for _, c := range cases {
		if ok := lim.AllowN("k", c.t, c.n); ok != c.ok {
			t.Errorf("AllowN(%v) = %v, want: %v", c.t, ok, c.ok)
		}
	}
}
