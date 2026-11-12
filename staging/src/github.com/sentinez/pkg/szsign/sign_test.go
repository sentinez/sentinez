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

package szsign

import (
	"bytes"
	"encoding/base64"
	"testing"

	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var (
	_key      = bytes.Repeat([]byte{1}, MinKeyBytes)
	_otherKey = bytes.Repeat([]byte{2}, MinKeyBytes)
)

func change() *centraldatapb.Change {
	return &centraldatapb.Change{
		Table:   "security.secrules",
		Op:      centraldatapb.Change_OP_UPSERT,
		Id:      "sr-1",
		Version: 7,
		Epoch:   "e1",
	}
}

func TestSealOpen(t *testing.T) {
	data, err := Seal(_key, change())
	require.NoError(t, err)

	got, err := Open(_key, data)

	require.NoError(t, err)
	assert.True(t, proto.Equal(change(), got))
}

func TestOpenWithAnotherKey(t *testing.T) {
	data, err := Seal(_key, change())
	require.NoError(t, err)

	_, err = Open(_otherKey, data)

	require.ErrorIs(t, err, ErrBadMAC)
}

func TestOpenAlteredChange(t *testing.T) {
	data, err := Seal(_key, change())
	require.NoError(t, err)

	var env centraldatapb.Envelope
	require.NoError(t, proto.Unmarshal(data, &env))
	forged := change()
	forged.Id = "sr-2"
	env.Change, err = proto.Marshal(forged)
	require.NoError(t, err)
	altered, err := proto.Marshal(&env)
	require.NoError(t, err)

	_, err = Open(_key, altered)

	require.ErrorIs(t, err, ErrBadMAC)
}

func TestOpenGarbage(t *testing.T) {
	_, err := Open(_key, []byte("not an envelope"))

	require.Error(t, err)
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		name    string
		give    string
		wantErr bool
	}{
		{"long enough", base64.StdEncoding.EncodeToString(_key), false},
		{"too short", base64.StdEncoding.EncodeToString([]byte("short")), true},
		{"not base64", "%%%", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseKey(tt.give)

			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
