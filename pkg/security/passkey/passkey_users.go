package passkey

import "github.com/go-webauthn/webauthn/webauthn"

type Users interface {
	webauthn.User
	AddCredential(*webauthn.Credential)
	UpdateCredential(*webauthn.Credential)
}
type User struct {
	ID          []byte
	DisplayName string
	Name        string

	cred []webauthn.Credential
}

func (u *User) WebAuthnID() []byte {
	return u.ID
}

func (u *User) WebAuthnName() string {
	return u.Name
}

func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *User) WebAuthnIcon() string {
	return "https://pics.com/avatar.png"
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.cred
}

func (u *User) AddCredential(credential *webauthn.Credential) {
	u.cred = append(u.cred, *credential)
}

func (u *User) UpdateCredential(credential *webauthn.Credential) {
	for i, c := range u.cred {
		if string(c.ID) == string(credential.ID) {
			u.cred[i] = *credential
		}
	}
}
