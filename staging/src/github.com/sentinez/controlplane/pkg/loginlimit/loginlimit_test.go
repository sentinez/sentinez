// Copyright 2026 Duc-Hung Ho.
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

package loginlimit

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// clock is a Limiter whose time the test moves.
func clock(limit int, window time.Duration) (*Limiter, *time.Time) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l := New(limit, window)
	l.now = func() time.Time { return now }

	return l, &now
}

func failN(l *Limiter, key string, n int) {
	for range n {
		l.Fail(key)
	}
}

func TestBlockedAfterMaxFailures(t *testing.T) {
	l, _ := clock(3, time.Minute)

	failN(l, "k", 2)
	assert.False(t, l.Blocked("k"))

	l.Fail("k")
	assert.True(t, l.Blocked("k"))
}

func TestKeysAreIndependent(t *testing.T) {
	l, _ := clock(2, time.Minute)

	failN(l, "a", 5)

	assert.True(t, l.Blocked("a"))
	assert.False(t, l.Blocked("b"))
}

func TestBlockEndsWithItsWindow(t *testing.T) {
	l, now := clock(2, time.Minute)
	failN(l, "k", 2)

	*now = now.Add(time.Minute)

	assert.False(t, l.Blocked("k"))
}

func TestMoreFailuresDoNotExtendTheBlock(t *testing.T) {
	l, now := clock(2, time.Minute)
	failN(l, "k", 2)

	*now = now.Add(50 * time.Second)
	failN(l, "k", 10) // an attacker keeps trying
	*now = now.Add(10 * time.Second)

	assert.False(t, l.Blocked("k"))
}

func TestFailuresAfterTheWindowStartOver(t *testing.T) {
	l, now := clock(2, time.Minute)
	l.Fail("k")

	*now = now.Add(2 * time.Minute)
	l.Fail("k")

	assert.False(t, l.Blocked("k"), "one failure in the new window")
}

func TestResetForgetsTheFailures(t *testing.T) {
	l, _ := clock(2, time.Minute)
	failN(l, "k", 2)

	l.Reset("k")

	assert.False(t, l.Blocked("k"))
}

func TestExpiredKeysAreForgotten(t *testing.T) {
	l, now := clock(2, time.Minute)
	failN(l, "old", 1)

	*now = now.Add(2 * time.Minute)
	l.Fail("new")

	assert.Len(t, l.entries, 1)
}

func TestKeysAreBounded(t *testing.T) {
	l, _ := clock(2, time.Hour)

	for i := range _maxKeys + 10 {
		l.Fail(strconv.Itoa(i))
	}

	assert.Len(t, l.entries, _maxKeys)
}
