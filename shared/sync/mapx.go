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

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

type Map[K comparable, V any] struct {
	core sync.Map
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	var empty V
	raw, ok := m.core.Load(key)
	if !ok {
		return empty, false
	}

	return raw.(V), true
}

func (m *Map[K, V]) Store(key K, value V) {
	m.core.Store(key, value)
}

func (m *Map[K, V]) Delete(key K) {
	m.core.Delete(key)
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.core.Range(func(k, v any) bool {
		typedKey, ok1 := k.(K)
		typedVal, ok2 := v.(V)
		if !ok1 || !ok2 {
			return true
		}

		return f(typedKey, typedVal)
	})
}

func (m *Map[K, V]) Keys() []K {
	var keys []K
	m.core.Range(func(k, _ any) bool {
		keys = append(keys, k.(K))
		return true
	})

	return keys
}

func (m *Map[K, V]) Values() []V {
	var values []V
	m.core.Range(func(_, v any) bool {
		values = append(values, v.(V))
		return true
	})

	return values
}

func (m *Map[K, V]) Clear() {
	m.core.Clear()
}
