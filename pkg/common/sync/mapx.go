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

package sync

import (
	"sync"
)

func New[K comparable, V any]() *MapX[K, V] {
	return &MapX[K, V]{}
}

type MapX[K comparable, V any] struct {
	core sync.Map
}

func (m *MapX[K, V]) Load(key K) (value V, ok bool) {
	var empty V
	raw, ok := m.core.Load(key)
	if !ok {
		return empty, false
	}

	return raw.(V), true
}

func (m *MapX[K, V]) Store(key K, value V) {
	m.core.Store(key, value)
}

func (m *MapX[K, V]) Delete(key K) {
	m.core.Delete(key)
}

func (m *MapX[K, V]) Range(f func(key K, value V) bool) {
	m.core.Range(func(k, v any) bool {
		typedKey, ok1 := k.(K)
		typedVal, ok2 := v.(V)
		if !ok1 || !ok2 {
			return true
		}

		return f(typedKey, typedVal)
	})
}

func (m *MapX[K, V]) Keys() []K {
	var keys []K
	m.core.Range(func(k, _ any) bool {
		keys = append(keys, k.(K))
		return true
	})

	return keys
}

func (m *MapX[K, V]) Values() []V {
	var values []V
	m.core.Range(func(_, v any) bool {
		values = append(values, v.(V))
		return true
	})

	return values
}

func (m *MapX[K, V]) Clear() {
	m.core.Clear()
}
