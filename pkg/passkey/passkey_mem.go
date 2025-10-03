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
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/sentinez/sentinez/pkg/infra/cache/mem"
	"github.com/sentinez/sentinez/pkg/uuid"
)

func NewMemoryStorage(ttl time.Duration) Store {
	return &MemoryStorage{
		ttl: ttl,

		// key: username
		users: mem.NewDefault[Users](),

		// key: token
		sessions: mem.NewDefault[*webauthn.SessionData](),
	}
}

type MemoryStorage struct {
	ttl      time.Duration
	users    *mem.Cache[Users]
	sessions *mem.Cache[*webauthn.SessionData]
}

// GenSessionID implements Store.
func (s *MemoryStorage) GenSessionID() (string, error) {
	return uuid.NewHex("PK"), nil
}

// DeleteSession implements Store.
func (s *MemoryStorage) DeleteSession(token string) {
	s.sessions.Del(token)
}

// GetSession implements Store.
func (s *MemoryStorage) GetSession(token string) (*webauthn.SessionData, bool) {
	ss, ok := s.sessions.Get(token)
	if !ok {
		return nil, ok
	}

	s.SaveSession(token, ss)

	return ss, false
}

// GetUser implements Store.
func (s *MemoryStorage) GetUser(userName string) Users {
	user, ok := s.users.Get(userName)
	if !ok {
		user = &User{
			ID:          []byte(userName),
			DisplayName: userName,
			Name:        userName,
		}
	}

	s.SaveUser(user)

	return user
}

// SaveSession implements Store.
func (s *MemoryStorage) SaveSession(token string, data *webauthn.SessionData) {
	s.sessions.SetWithTTL(token, data, s.ttl)
}

// SaveUser implements Store.
func (s *MemoryStorage) SaveUser(user Users) {
	s.users.SetWithTTL(user.WebAuthnName(), user, s.ttl)
}
