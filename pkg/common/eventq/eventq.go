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

package eventq

import (
	"sync"
)

// New creates a new queue channel with size
func New[T any](size int) *QueueChannel[T] {
	return &QueueChannel[T]{Size: size}
}

// QueueChannel is a queue with size and mapping of names to channel
type QueueChannel[T any] struct {
	Size     int
	channels sync.Map
}

// Get channel queue with names key, queue supports string type only
func (q *QueueChannel[T]) Get(namespace string) chan T {
	if q == nil {
		return nil
	}

	ch, _ := q.channels.LoadOrStore(namespace, make(chan T, q.Size))

	return ch.(chan T)
}

// Close channel queue with names key
func (q *QueueChannel[T]) Close(namespace string) {
	if q == nil {
		return
	}

	ch, ok := q.channels.Load(namespace)
	if ok {
		close(ch.(chan string))
		q.channels.Delete(namespace)
	}
}

// CloseAll closes all channels in the queue
func (q *QueueChannel[T]) CloseAll() {
	if q == nil {
		return
	}

	q.channels.Range(func(key, value any) bool {
		close(value.(chan any))
		q.channels.Delete(key)
		return true
	})
}
