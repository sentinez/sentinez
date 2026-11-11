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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/sentinez/pkg/szconfig"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
)

// _sealedPrefix tells the format of a sealed value, so a later format
// (another key, another algorithm) can be told apart.
const _sealedPrefix = "v1:"

// _keySize is the size in bytes of an AES-256 key.
const _keySize = 32

var (
	// ErrNoDataKey SENZ_DATA_KEY is not set.
	ErrNoDataKey = errors.New("crypto: SENZ_DATA_KEY is empty")

	// ErrOpen a sealed value can not be opened: it is malformed, was
	// altered, belongs to another record or was sealed with another key.
	ErrOpen = errors.New("crypto: open sealed value")
)

// Sealer encrypts secrets before they are stored, with AES-256-GCM.
type Sealer struct {
	aead cipher.AEAD
}

// NewSealer returns a Sealer for key, 32 bytes encoded as base64.
func NewSealer(key string) (*Sealer, error) {
	raw, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: data key is not base64: %w", err)
	}

	if len(raw) != _keySize {
		return nil, fmt.Errorf("crypto: data key must be %d bytes, got %d",
			_keySize, len(raw))
	}

	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}

	return &Sealer{aead: aead}, nil
}

// NewSealerFromEnv returns the Sealer of SENZ_DATA_KEY, ErrNoDataKey when it
// is not set: secrets are never stored in clear because of a missing key.
func NewSealerFromEnv() (*Sealer, error) {
	key := szconfig.GetEnv(settingpb.Senz_SENZ_DATA_KEY)
	if key == "" {
		return nil, ErrNoDataKey
	}

	return NewSealer(key)
}

// Seal encrypts plain. aad is authenticated but not stored, give the id of
// the record: the value then only opens for that record.
func (s *Sealer) Seal(plain, aad string) (string, error) {
	nonce := make([]byte, s.aead.NonceSize(), s.aead.NonceSize()+
		len(plain)+s.aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("crypto: nonce: %w", err)
	}

	sealed := s.aead.Seal(nonce, nonce, []byte(plain), []byte(aad))

	return _sealedPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// Open decrypts a value of Seal with the same aad, ErrOpen when it can not.
func (s *Sealer) Open(sealed, aad string) (string, error) {
	encoded, ok := strings.CutPrefix(sealed, _sealedPrefix)
	if !ok {
		return "", ErrOpen
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) < s.aead.NonceSize()+s.aead.Overhead() {
		return "", ErrOpen
	}

	nonce, body := raw[:s.aead.NonceSize()], raw[s.aead.NonceSize():]

	plain, err := s.aead.Open(nil, nonce, body, []byte(aad))
	if err != nil {
		return "", ErrOpen
	}

	return string(plain), nil
}
