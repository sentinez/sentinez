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

package limiter

import "time"

type Window interface {
	// Start returns the start boundary.
	Start() time.Time

	// Count returns the accumulated count.
	Count() int64

	// AddCount increments the accumulated count by n.
	AddCount(n int64)

	// Reset sets the state of the window with the given settings.
	Reset(s time.Time, c int64)

	// Sync tries to exchange data between the window and the central
	// datastore at time now, to keep the window's count up-to-date.
	Sync(now time.Time)
}

var _ Window = (*LocalWindow)(nil)

func NewLocalWindow() *LocalWindow {
	return &LocalWindow{}
}

type LocalWindow struct {
	// The start boundary (timestamp in nanoseconds) of the window
	// [start, start + windowSize (time)]
	start int64

	// total count
	count int64
}

func (lw *LocalWindow) Start() time.Time {
	return time.Unix(0, lw.start)
}

func (lw *LocalWindow) Count() int64 {
	return lw.count
}

func (lw *LocalWindow) AddCount(n int64) {
	lw.count += n
}

func (lw *LocalWindow) Reset(s time.Time, c int64) {
	lw.start = s.UnixNano()
	lw.count = c
}

func (lw *LocalWindow) Sync(_ time.Time) {}
