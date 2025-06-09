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

// Package decor provides a simple way to decorate functions
package decor

import "context"

// WithContextReturn executes the function fn and returns its result.
func WithContextReturn[T any](
	ctx context.Context, fn func() (T, error)) (T, error) {
	var zero T
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	default:
		return fn()
	}
}

// WithContext executes the function fn and returns its error.
func WithContext(ctx context.Context, fn func() error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fn()
	}
}
