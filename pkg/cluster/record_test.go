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

package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordCodec(t *testing.T) {
	tests := []struct {
		name string
		give record
	}{
		{
			name: "value",
			give: record{version: 42, origin: "edge-a", payload: []byte("v")},
		},
		{
			name: "tombstone",
			give: record{version: 7, origin: "edge-b", deleted: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, got, err := decodeMessage(encodeMessage("k", tt.give))

			require.NoError(t, err)
			assert.Equal(t, "k", key)
			assert.Equal(t, tt.give.version, got.version)
			assert.Equal(t, tt.give.origin, got.origin)
			assert.Equal(t, tt.give.deleted, got.deleted)
			assert.Equal(t, string(tt.give.payload), string(got.payload))
		})
	}
}

func TestDecodeRecordMalformed(t *testing.T) {
	tests := []struct {
		name string
		give []byte
	}{
		{name: "short", give: []byte{1, 2}},
		{name: "origin past end", give: append(make([]byte, 9), 50)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeRecord(tt.give)

			require.ErrorIs(t, err, errBadRecord)
		})
	}
}

func TestRecordNewer(t *testing.T) {
	tests := []struct {
		name string
		give record
		than record
		want bool
	}{
		{
			name: "higher version",
			give: record{version: 2, origin: "a"},
			than: record{version: 1, origin: "z"},
			want: true,
		},
		{
			name: "lower version",
			give: record{version: 1, origin: "z"},
			than: record{version: 2, origin: "a"},
		},
		{
			name: "tie broken by origin",
			give: record{version: 1, origin: "b"},
			than: record{version: 1, origin: "a"},
			want: true,
		},
		{
			name: "same write",
			give: record{version: 1, origin: "a"},
			than: record{version: 1, origin: "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.newer(tt.than))
		})
	}
}

func TestHLC(t *testing.T) {
	c := newHLC()
	c.now = func() uint64 { return 100 }

	assert.Equal(t, uint64(100), c.next())
	assert.Equal(t, uint64(101), c.next(), "never goes back")

	c.observe(500)
	assert.Equal(t, uint64(501), c.next(), "moves past what it saw")
}
