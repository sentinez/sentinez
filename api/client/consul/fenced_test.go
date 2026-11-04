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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _fencedKey = "sentinez/dmz/edge/zone/sgn"

func TestPutFenced(t *testing.T) {
	tests := []struct {
		name      string
		giveToken uint64
	}{
		{name: "same writer", giveToken: 5},
		{name: "newer writer", giveToken: 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			c := newTestClient(t)
			require.NoError(t, c.PutFenced(ctx, _fencedKey, 5, "a"))

			require.NoError(t,
				c.PutFenced(ctx, _fencedKey, tt.giveToken, "b"))

			got, err := GetFenced[string](ctx, c, _fencedKey)
			require.NoError(t, err)
			assert.Equal(t, tt.giveToken, got.Token)
			assert.Equal(t, "b", got.Value)
		})
	}
}

func TestPutFencedRejectsLowerToken(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)
	require.NoError(t, c.PutFenced(ctx, _fencedKey, 5, "a"))

	err := c.PutFenced(ctx, _fencedKey, 4, "stale")

	require.ErrorIs(t, err, ErrFenced)
	got, err := GetFenced[string](ctx, c, _fencedKey)
	require.NoError(t, err)
	assert.Equal(t, "a", got.Value)
}

func TestPutFencedCASConflicts(t *testing.T) {
	tests := []struct {
		name     string
		giveMiss int
		wantErr  error
	}{
		{name: "retried", giveMiss: _casRetries - 1},
		{name: "gives up", giveMiss: _casRetries, wantErr: errCASConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, kv := newTestKV(t)
			kv.casMiss = tt.giveMiss

			err := c.PutFenced(context.Background(), _fencedKey, 1, "a")

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestGetFencedNotFound(t *testing.T) {
	_, err := GetFenced[string](context.Background(),
		newTestClient(t), _fencedKey)

	require.ErrorIs(t, err, ErrKeyNotFound)
}
