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

// Package mem implements a memory cache
package mem

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// New create instance memory cache
//
// Parameters:
//   - defaultExpr is a time to live of items
//   - cleanInterval is cleanup cycle for expired items
func New[T any](
	defaultExpr time.Duration, cleanInterval time.Duration) *Cache[T] {

	return &Cache[T]{
		core: cache.New(defaultExpr, cleanInterval),
	}
}

// NewDefault create new memory cache with, default expr is 1 minute
// and cleanup cycle is 1 minute
func NewDefault[T any]() *Cache[T] {
	return &Cache[T]{
		core: cache.New(time.Minute, time.Minute),
	}
}

// Cache implement a memory cache
type Cache[T any] struct {
	core *cache.Cache
}

// Set used to set key with value into memory
func (c *Cache[T]) Set(k string, v T) {
	c.core.Set(k, v, cache.DefaultExpiration)
}

// SetWithTTL used to set key with value with time to live
func (c *Cache[T]) SetWithTTL(k string, v T, ttl time.Duration) {
	c.core.Set(k, v, ttl)
}

// Get used to get value of key
//
// Returns:
//   - result T is result data of Get function
//   - found bool is a second value of Get,
//     true when value is found,
//     false when value is not found
func (c *Cache[T]) Get(k string) (T, bool) {
	result, found := c.core.Get(k)
	if !found {
		var emptyValue T
		return emptyValue, false
	}

	return result.(T), true
}

// Del used to delete key from memory
func (c *Cache[T]) Del(k string) {
	c.core.Delete(k)
}

// List used to list all items in cache
func (c *Cache[T]) List() ([]T, bool) {
	var results []T

	for _, val := range c.core.Items() {
		obj, ok := val.Object.(T)
		if !ok {
			return nil, false
		}

		results = append(results, obj)
	}

	return results, true
}
