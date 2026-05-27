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
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/iam/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	"github.com/sentinez/shared/zlog"
)

// nolint
type Store interface {
	GetSession(token string) (*webauthn.SessionData, bool)
	SaveSession(token string, data *webauthn.SessionData)
	DeleteSession(token string)
	GenSessionID() (string, error)
	GetOrCreateAccount(email string) (*iampb.Account, error)
	GetAndDeleteAccount(email string) (*iampb.Account, error)
}

func NewWebAuthn(config *confpb.Config) *webauthn.WebAuthn {
	wconfig := &webauthn.Config{
		// Display Name for your site
		RPDisplayName: core.Name,
		// Generally the FQDN for your site
		RPID: config.GetEnv().GetHostname(),
		// The origin URLs allowed for WebAuthn
		RPOrigins: []string{config.GetEnv().GetClientOrigin()},
	}

	wauth, err := webauthn.New(wconfig)
	if err != nil {
		zlog.Errorf("[PASSKEY] error when create webauthn: %v", err)
	}

	return wauth
}
