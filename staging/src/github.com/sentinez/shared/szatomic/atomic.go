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

package szatomic

import (
	"sync/atomic"

	"github.com/sentinez/shared/szlog"
)

type WriteOnce[T any] struct {
	value atomic.Pointer[T]
}

func (w *WriteOnce[T]) Set(v T) bool {
	p := new(T)
	*p = v

	if !w.value.CompareAndSwap(nil, p) {
		szlog.Warn("atomic: value has already been set")
		return false
	}

	return true
}

func (w *WriteOnce[T]) Get() (T, bool) {
	p := w.value.Load()
	if p == nil {
		var zero T
		return zero, false
	}

	return *p, true
}

func (w *WriteOnce[T]) Reset() {
	w.value.Store(nil)
}
