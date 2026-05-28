// Copyright 2025 Sentinéz Labs.
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

package accrepos

import (
	"github.com/go-webauthn/webauthn/webauthn"
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/iam/v1"
	"github.com/sentinez/shared/jsonx"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/proto"
)

type IAccountX interface {
	proto.Message

	webauthn.User
	AddCredential(*webauthn.Credential)
	UpdateCredential(*webauthn.Credential)
}

type AccountX struct {
	*iampb.Account
}

func (ax *AccountX) AddCredential(credential *webauthn.Credential) {
	data, err := jsonx.Marshal(credential)
	if err != nil {
		zlog.Errorf("failed to marshal credential: %v", err)
	}
	ax.Credentials = append(ax.Credentials, string(data))
}

func (ax *AccountX) UpdateCredential(credential *webauthn.Credential) {
	for i, c := range ax.GetCredentials() {
		var cred webauthn.Credential

		if err := jsonx.Unmarshal([]byte(c), &cred); err != nil {
			zlog.Errorf("failed to unmarshal credential: %v", err)
			return
		}

		if string(cred.ID) == string(credential.ID) {
			data, err := jsonx.Marshal(credential)
			if err != nil {
				zlog.Errorf("failed to marshal credential: %v", err)
				return
			}

			ax.Credentials[i] = string(data)
		}
	}
}

func (ax *AccountX) WebAuthnID() []byte {
	return []byte(ax.GetEmail())
}

func (ax *AccountX) WebAuthnName() string {
	return ax.GetEmail()
}

func (ax *AccountX) WebAuthnDisplayName() string {
	return ax.GetEmail()
}

func (ax *AccountX) WebAuthnCredentials() []webauthn.Credential {
	var resp []webauthn.Credential
	for _, c := range ax.GetCredentials() {
		var cred webauthn.Credential
		if err := jsonx.Unmarshal([]byte(c), &cred); err != nil {
			zlog.Errorf("failed to unmarshal credential: %v", err)
			continue
		}

		resp = append(resp, cred)
	}

	return resp
}
