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

package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keyOf(b byte, size int) string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{b}, size))
}

func newTestSealer(t *testing.T, b byte) *Sealer {
	t.Helper()

	s, err := NewSealer(keyOf(b, 32))
	require.NoError(t, err)

	return s
}

func TestNewSealerRejectsBadKeys(t *testing.T) {
	tests := []struct {
		name    string
		giveKey string
	}{
		{"empty", ""},
		{"not base64", "%%%"},
		{"16 bytes", keyOf(1, 16)},
		{"33 bytes", keyOf(1, 33)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSealer(tt.giveKey)

			assert.Error(t, err)
		})
	}
}

func TestSealOpenRoundTrip(t *testing.T) {
	s := newTestSealer(t, 1)

	sealed, err := s.Seal("-----BEGIN PRIVATE KEY-----", "cacrt_1")
	require.NoError(t, err)

	plain, err := s.Open(sealed, "cacrt_1")

	require.NoError(t, err)
	assert.Equal(t, "-----BEGIN PRIVATE KEY-----", plain)
	assert.NotContains(t, sealed, "PRIVATE")
	assert.True(t, len(sealed) > len(_sealedPrefix))
}

func TestSealIsRandomized(t *testing.T) {
	s := newTestSealer(t, 1)

	first, err := s.Seal("secret", "cacrt_1")
	require.NoError(t, err)
	second, err := s.Seal("secret", "cacrt_1")
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

func TestOpenFails(t *testing.T) {
	s := newTestSealer(t, 1)
	sealed, err := s.Seal("secret", "cacrt_1")
	require.NoError(t, err)

	raw, err := base64.StdEncoding.DecodeString(sealed[len(_sealedPrefix):])
	require.NoError(t, err)
	raw[len(raw)-1] ^= 0x01
	altered := _sealedPrefix + base64.StdEncoding.EncodeToString(raw)

	tests := []struct {
		name       string
		giveSealer *Sealer
		giveSealed string
		giveAad    string
	}{
		{"another record", s, sealed, "cacrt_2"},
		{"another key", newTestSealer(t, 2), sealed, "cacrt_1"},
		{"altered", s, altered, "cacrt_1"},
		{"no prefix", s, sealed[len(_sealedPrefix):], "cacrt_1"},
		{"not base64", s, _sealedPrefix + "%%%", "cacrt_1"},
		{"too short", s, _sealedPrefix + "AAAA", "cacrt_1"},
		{"empty", s, "", "cacrt_1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plain, err := tt.giveSealer.Open(tt.giveSealed, tt.giveAad)

			assert.ErrorIs(t, err, ErrOpen)
			assert.Empty(t, plain)
		})
	}
}
