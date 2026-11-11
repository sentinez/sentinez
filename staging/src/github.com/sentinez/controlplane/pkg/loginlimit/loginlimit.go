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

// Package loginlimit limits the failed sign-in attempts.
package loginlimit

import (
	"sync"
	"time"
)

// _maxKeys bounds the memory: the keys come from the callers, who pick
// them. Past it a new key is not counted.
const _maxKeys = 100_000

// Limiter blocks a key after limit failures in a window. Only the failures
// count, so a user who signs in is never held back, and a block ends with
// its window: more attempts do not extend it.
//
// It forgets the expired keys as it is used, with no goroutine to stop.
type Limiter struct {
	limit  int
	window time.Duration
	now    func() time.Time

	mu      sync.Mutex
	entries map[string]*entry
	sweepAt time.Time
}

type entry struct {
	failures int
	// start is when the window of the failures began.
	start time.Time
}

// New returns a Limiter that blocks a key after limit failures in window.
func New(limit int, window time.Duration) *Limiter {
	return &Limiter{
		limit:   limit,
		window:  window,
		now:     time.Now,
		entries: make(map[string]*entry),
	}
}

// Blocked reports whether key had limit failures in the current window.
func (l *Limiter) Blocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[key]

	return ok && !l.expired(e, l.now()) && e.failures >= l.limit
}

// Fail counts a failure of key.
func (l *Limiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	e, ok := l.entries[key]
	if !ok || l.expired(e, now) {
		if !ok && len(l.entries) >= _maxKeys {
			return
		}

		e = &entry{start: now}
		l.entries[key] = e
	}

	e.failures++
}

// Reset forgets the failures of key, after a sign-in.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.entries, key)
}

func (l *Limiter) expired(e *entry, now time.Time) bool {
	return now.Sub(e.start) >= l.window
}

// sweep drops the expired keys, at most once per window.
func (l *Limiter) sweep(now time.Time) {
	if now.Before(l.sweepAt) {
		return
	}

	for key, e := range l.entries {
		if l.expired(e, now) {
			delete(l.entries, key)
		}
	}

	l.sweepAt = now.Add(l.window)
}
