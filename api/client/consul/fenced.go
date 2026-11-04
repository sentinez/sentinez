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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"
)

// _casRetries bounds the compare-and-swap attempts of PutFenced when
// other writers keep changing the key in between.
const _casRetries = 5

var (
	// ErrFenced is returned when a write carries a lower fencing token
	// than the one stored: its writer has been deposed.
	ErrFenced = errors.New("consul: fenced")

	errCASConflict = errors.New("consul: compare-and-swap conflict")
)

// Fenced is a value stored with the fencing token of its writer.
type Fenced[T any] struct {
	Token uint64 `json:"token"`
	Value T      `json:"value"`
}

// PutFenced stores value under key with token, as JSON, unless the key
// holds a higher token (ErrFenced); an equal token is the same writer.
// The token check and the write are atomic: the write is a CAS on the
// ModifyIndex the check read, served by Consul's Raft log.
func (c *Client) PutFenced(
	ctx context.Context, key string, token uint64, value any,
) error {
	// always JSON: the token is read back whatever the key extension
	raw, err := encode(Fenced[any]{Token: token, Value: value}, FormatJSON)
	if err != nil {
		return fmt.Errorf("encode %q: %w", key, err)
	}

	for range _casRetries {
		ok, err := c.casFenced(ctx, key, token, raw)
		if err != nil || ok {
			return err
		}
	}

	return fmt.Errorf("put fenced %q: %w", key, errCASConflict)
}

// GetFenced returns the value under key and the token of its writer.
func GetFenced[T any](
	ctx context.Context, c *Client, key string,
) (Fenced[T], error) {
	return Get[Fenced[T]](ctx, c, key, WithFormat(FormatJSON))
}

// casFenced checks the token stored under key, then writes raw if the
// key is unchanged since; false when another write came in between.
func (c *Client) casFenced(
	ctx context.Context, key string, token uint64, raw []byte,
) (bool, error) {
	opts := (&api.QueryOptions{RequireConsistent: true}).WithContext(ctx)

	pair, _, err := c.consul.KV().Get(key, opts)
	if err != nil {
		return false, fmt.Errorf("get %q: %w", key, err)
	}

	// ModifyIndex 0 makes the CAS create the key only if it is missing
	var index uint64
	if pair != nil {
		var stored Fenced[json.RawMessage]
		if err := json.Unmarshal(pair.Value, &stored); err != nil {
			return false, fmt.Errorf("decode %q: %w", key, err)
		}

		if stored.Token > token {
			return false, fmt.Errorf("put fenced %q: token %d < %d: %w",
				key, token, stored.Token, ErrFenced)
		}

		index = pair.ModifyIndex
	}

	wopts := (&api.WriteOptions{}).WithContext(ctx)
	next := &api.KVPair{Key: key, Value: raw, ModifyIndex: index}

	ok, _, err := c.consul.KV().CAS(next, wopts)
	if err != nil {
		return false, fmt.Errorf("cas %q: %w", key, err)
	}

	return ok, nil
}
