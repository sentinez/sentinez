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
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/sentinez/core"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/shared/zlog"
)

// Registration is a passkey registration that waits for the authenticator
// response. Nothing is persisted until that response is verified.
type Registration struct {
	// UserId is the user the passkey is added to. It is empty when the
	// registration signs up a new user.
	UserId string
	Email  string
	// Handle is the WebAuthn user handle given to the authenticator.
	Handle string
}

// nolint
type Store interface {
	GetSession(token string) (*webauthn.SessionData, bool)
	SaveSession(token string, data *webauthn.SessionData)
	DeleteSession(token string)
	GenSessionId() (string, error)
	SaveRegistration(token string, reg *Registration)
	// TakeRegistration returns the registration saved under token and
	// forgets it, so a registration is finished at most once.
	TakeRegistration(token string) (*Registration, bool)
}

func NewWebAuthn(config *settingpb.Config) *webauthn.WebAuthn {
	wconfig := &webauthn.Config{
		// Display Name for your site
		RPDisplayName: core.Name,
		// Generally the FQDN for your site
		RPID: config.Get(settingpb.Senz_SENZ_HOSTNAME),
		// The origin URLs allowed for WebAuthn
		RPOrigins: []string{config.Get(settingpb.Senz_SENZ_CLIENT_ORIGIN)},
	}

	wauth, err := webauthn.New(wconfig)
	if err != nil {
		zlog.Errorf("[PASSKEY] error when create webauthn: %v", err)
	}

	return wauth
}
