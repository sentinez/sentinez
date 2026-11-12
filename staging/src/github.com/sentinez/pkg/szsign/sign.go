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

// Package szsign signs the changes centraldata publishes to the edges: the
// channel they travel on does not tell who published, so an edge applies a
// change only when its HMAC proves the sender holds the sync key.
package szsign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	centraldatapb "github.com/sentinez/sentinez/api/proto/sentinez/apps/centraldata/v1"
	"google.golang.org/protobuf/proto"
)

// MinKeyBytes the shortest key accepted.
const MinKeyBytes = 32

// ErrBadMAC the envelope was not signed with the key.
var ErrBadMAC = errors.New("szsign: bad mac")

// ParseKey decodes the base64 key (SENZ_SYNC_KEY), which must hold at
// least [MinKeyBytes] bytes.
func ParseKey(encoded string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("szsign: key: %w", err)
	}

	if len(key) < MinKeyBytes {
		return nil, fmt.Errorf("szsign: key is shorter than %d bytes",
			MinKeyBytes)
	}

	return key, nil
}

// Seal returns the marshaled envelope of change signed with key.
func Seal(key []byte, change *centraldatapb.Change) ([]byte, error) {
	body, err := proto.MarshalOptions{Deterministic: true}.Marshal(change)
	if err != nil {
		return nil, fmt.Errorf("szsign: marshal change: %w", err)
	}

	return proto.Marshal(&centraldatapb.Envelope{
		Change: body,
		Mac:    mac(key, body),
	})
}

// Open verifies the marshaled envelope with key and returns its change;
// [ErrBadMAC] when it was signed with another key or was altered.
func Open(key, data []byte) (*centraldatapb.Change, error) {
	var env centraldatapb.Envelope
	if err := proto.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("szsign: unmarshal envelope: %w", err)
	}

	if !hmac.Equal(env.GetMac(), mac(key, env.GetChange())) {
		return nil, ErrBadMAC
	}

	var change centraldatapb.Change
	if err := proto.Unmarshal(env.GetChange(), &change); err != nil {
		return nil, fmt.Errorf("szsign: unmarshal change: %w", err)
	}

	return &change, nil
}

func mac(key, body []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(body)

	return h.Sum(nil)
}
