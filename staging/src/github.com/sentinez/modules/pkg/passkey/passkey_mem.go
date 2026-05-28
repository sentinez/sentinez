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

package passkey

import (
	"net/mail"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/sentinez/core"
	"github.com/sentinez/core/storage/cache/mem"
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/iam/v1"
	"github.com/sentinez/shared/errorx"
	"github.com/sentinez/shared/rand"
)

func NewMemoryStorage(ttl time.Duration) Store {
	return &MemoryStorage{
		ttl: ttl,

		// key: token
		sessions: mem.NewDefault[*webauthn.SessionData](),

		// key: email
		accounts: mem.NewDefault[*iampb.Account](),
	}
}

type MemoryStorage struct {
	ttl      time.Duration
	sessions *mem.Cache[*webauthn.SessionData]
	accounts *mem.Cache[*iampb.Account]
}

// GetAndDeleteAccount implements Store.
func (s *MemoryStorage) GetAndDeleteAccount(
	email string) (*iampb.Account, error) {

	acc, ok := s.accounts.Get(email)
	if ok {
		s.accounts.Del(email)
		return acc, nil
	}

	return nil, errorx.ErrNotFound
}

// GetOrCreateAccount implements Store.
func (s *MemoryStorage) GetOrCreateAccount(
	email string) (*iampb.Account, error) {

	acc, ok := s.accounts.Get(email)
	if ok {
		return acc, nil
	}

	acc = &iampb.Account{Email: email, Username: email}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, err
	}

	s.accounts.Set(email, acc)

	return acc, nil
}

// GenSessionID implements Store.
func (s *MemoryStorage) GenSessionID() (string, error) {
	return rand.NewNanoID(core.Code + "SS"), nil
}

// DeleteSession implements Store.
func (s *MemoryStorage) DeleteSession(token string) {
	s.sessions.Del(token)
}

// GetSession implements Store.
func (s *MemoryStorage) GetSession(token string) (*webauthn.SessionData, bool) {
	ss, ok := s.sessions.Get(token)
	if !ok {
		return nil, false
	}

	s.SaveSession(token, ss)

	return ss, true
}

// SaveSession implements Store.
func (s *MemoryStorage) SaveSession(token string, ss *webauthn.SessionData) {
	s.sessions.SetWithTTL(token, ss, s.ttl)
}
