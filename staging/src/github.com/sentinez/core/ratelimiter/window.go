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

type LocalWindow struct{}

func (lw *LocalWindow) Start() time.Time {
	//TODO implement me
	panic("implement me")
}

func (lw *LocalWindow) Count() int64 {
	//TODO implement me
	panic("implement me")
}

func (lw *LocalWindow) AddCount(n int64) {
	//TODO implement me
	panic("implement me")
}

func (lw *LocalWindow) Reset(s time.Time, c int64) {
	//TODO implement me
	panic("implement me")
}

func (lw *LocalWindow) Sync(now time.Time) {
	//TODO implement me
	panic("implement me")
}
