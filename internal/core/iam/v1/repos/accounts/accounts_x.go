// Copyright 2025 Sentinez Labs.
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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/pkg/common/copier"
	"github.com/sentinez/sentinez/pkg/common/protobuf/protox"
	"google.golang.org/protobuf/proto"
)

type IAccountX interface {
	proto.Message

	webauthn.User
	AddCredential(*webauthn.Credential)
	UpdateCredential(*webauthn.Credential)
}

type AccountX struct {
	*iam.Account
}

func (ax *AccountX) AddCredential(credential *webauthn.Credential) {
	cred := protox.Struct(credential)
	ax.Credentials = append(ax.Credentials, cred)
}

func (ax *AccountX) UpdateCredential(credential *webauthn.Credential) {
	var creds []webauthn.Credential
	_ = copier.CopyJSON(ax.Credentials, &creds)
	for i, c := range creds {
		if string(c.ID) == string(credential.ID) {
			creds[i] = *credential
		}
	}

	_ = copier.CopyJSON(&creds, ax.Credentials)
}

func (ax *AccountX) WebAuthnID() []byte {
	return []byte(ax.Id)
}

func (ax *AccountX) WebAuthnName() string {
	return ax.GetUserId()
}

func (ax *AccountX) WebAuthnDisplayName() string {
	return ax.Username
}

func (ax *AccountX) WebAuthnCredentials() []webauthn.Credential {
	var resp []webauthn.Credential
	_ = copier.CopyJSON(ax.Credentials, &resp)
	return resp
}
