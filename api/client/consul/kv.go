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

package consul

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
)

// ErrKeyNotFound is returned when the requested key does not exist.
var ErrKeyNotFound = errors.New("consul: key not found")

// GetRaw returns the raw value stored under key.
func (c *Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	opts := (&api.QueryOptions{}).WithContext(ctx)
	pair, _, err := c.consul.KV().Get(key, opts)
	if err != nil {
		return nil, fmt.Errorf("get %q: %w", key, err)
	}
	if pair == nil {
		return nil, ErrKeyNotFound
	}
	return pair.Value, nil
}

// Put stores value under key. A string or []byte is stored as is; any
// other value is encoded in the format resolved from opts and key (see
// WithFormat).
func (c *Client) Put(
	ctx context.Context,
	key string,
	value any,
	opts ...Option,
) error {
	raw, err := encode(value, resolveFormat(key, opts))
	if err != nil {
		return fmt.Errorf("encode %q: %w", key, err)
	}

	pair := &api.KVPair{Key: key, Value: raw}
	wopts := (&api.WriteOptions{}).WithContext(ctx)
	if _, err := c.consul.KV().Put(pair, wopts); err != nil {
		return fmt.Errorf("put %q: %w", key, err)
	}
	return nil
}

// Delete removes key. Deleting a missing key is not an error.
func (c *Client) Delete(ctx context.Context, key string) error {
	opts := (&api.WriteOptions{}).WithContext(ctx)
	if _, err := c.consul.KV().Delete(key, opts); err != nil {
		return fmt.Errorf("delete %q: %w", key, err)
	}
	return nil
}

// Get returns the value under key decoded as T. A string or []byte T
// receives the raw value; any other T is decoded in the format resolved
// from opts and key (see WithFormat).
func Get[T any](
	ctx context.Context,
	c *Client,
	key string,
	opts ...Option,
) (T, error) {
	var v T
	raw, err := c.GetRaw(ctx, key)
	if err != nil {
		return v, err
	}
	if err := decode(raw, &v, resolveFormat(key, opts)); err != nil {
		return v, fmt.Errorf("decode %q: %w", key, err)
	}
	return v, nil
}

// GetOr returns the value under key decoded as T, or def when the key
// does not exist.
func GetOr[T any](
	ctx context.Context,
	c *Client,
	key string,
	def T,
	opts ...Option,
) (T, error) {
	v, err := Get[T](ctx, c, key, opts...)
	if errors.Is(err, ErrKeyNotFound) {
		return def, nil
	}
	return v, err
}

// List returns every value under prefix decoded as T, keyed by the key
// relative to prefix. Folder entries (keys ending with "/") are skipped.
// The format is resolved per key (see WithFormat).
func List[T any](
	ctx context.Context,
	c *Client,
	prefix string,
	opts ...Option,
) (map[string]T, error) {
	qopts := (&api.QueryOptions{}).WithContext(ctx)
	pairs, _, err := c.consul.KV().List(prefix, qopts)
	if err != nil {
		return nil, fmt.Errorf("list %q: %w", prefix, err)
	}

	result := make(map[string]T, len(pairs))
	for _, pair := range pairs {
		if strings.HasSuffix(pair.Key, "/") {
			continue
		}
		var v T
		f := resolveFormat(pair.Key, opts)
		if err := decode(pair.Value, &v, f); err != nil {
			return nil, fmt.Errorf("decode %q: %w", pair.Key, err)
		}
		result[strings.TrimPrefix(pair.Key, prefix)] = v
	}
	return result, nil
}
