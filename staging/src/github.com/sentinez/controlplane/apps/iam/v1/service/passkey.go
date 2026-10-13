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

package iamsvc

import (
	"context"
	"encoding/json"
	"errors"
	"net/mail"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	accrepos "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts"
	"github.com/sentinez/controlplane/pkg/headers"
	"github.com/sentinez/controlplane/pkg/passkey"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	"github.com/sentinez/shared/errorx"
	"github.com/sentinez/shared/protobuf/protox"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/structpb"
)

// PasskeyLoginVerify implements iampb.IdentityAccessManagementServiceServer.
func (srv *IAMService) PasskeyLoginVerify(ctx context.Context,
	req *iampb.PasskeyLoginVerifyRequest,
) (*iampb.PasskeyLoginVerifyResponse, error) {

	sid := req.GetSessionId()
	ss, ok := srv.store.GetSession(sid)
	if !ok {
		return nil, errorx.StatusNotFoundF("session not found id=%s", sid)
	}

	// ss.UserID is the handle returned by AccountX.WebAuthnID
	acc, err := srv.accounts.GetByProviderUserId(
		ctx, iampb.Provider_PROVIDER_PASSKEY, string(ss.UserID))
	if err != nil {
		return nil, err
	}

	credential, err := srv.validateLogin(
		acc, ss, req.GetCredentialAssertionData())
	if err != nil {
		return nil, err
	}

	acc.UpdateCredential(credential)
	if err = srv.accounts.Update(ctx, acc); err != nil {
		return nil, err
	}

	srv.store.DeleteSession(sid)

	return srv.passkeyLoginResponse(ctx, acc)
}

func (srv *IAMService) validateLogin(acc *accrepos.AccountX,
	ss *webauthn.SessionData, data []byte) (*webauthn.Credential, error) {

	var car protocol.CredentialAssertionResponse
	if err := json.Unmarshal(data, &car); err != nil {
		return nil, err
	}

	parsedCAR, err := car.Parse()
	if err != nil {
		return nil, err
	}

	credential, err := srv.auth.ValidateLogin(acc, *ss, parsedCAR)
	if err != nil {
		return nil, err
	}

	if credential.Authenticator.CloneWarning {
		zlog.Warnf("can't finish login: %s", "CloneWarning")
	}

	return credential, nil
}

func (srv *IAMService) passkeyLoginResponse(ctx context.Context,
	acc *accrepos.AccountX) (*iampb.PasskeyLoginVerifyResponse, error) {

	user, err := srv.users.Get(ctx, acc.GetUserId())
	if err != nil {
		return nil, err
	}

	accessToken, err := newAccessToken(user)
	if err != nil {
		return nil, err
	}

	return &iampb.PasskeyLoginVerifyResponse{
		AccessToken: accessToken,
		User:        user,
	}, nil
}

// PasskeyLoginChallenge implements iampb.IdentityAccessManagementServiceServer.
func (srv *IAMService) PasskeyLoginChallenge(
	ctx context.Context,
	req *iampb.PasskeyLoginChallengeRequest,
) (*iampb.PasskeyLoginChallengeResponse, error) {

	emailOrUsername := req.GetEmailOrUsername()
	acc, err := srv.accounts.GetByProvider(
		ctx, iampb.Provider_PROVIDER_PASSKEY, emailOrUsername)
	if errors.Is(err, errorx.ErrNotFound) {
		return nil, errorx.StatusNotFoundF(
			"not found username or email=%s", emailOrUsername)
	}

	if err != nil {
		return nil, err
	}

	options, session, err := srv.auth.BeginLogin(acc)
	if err != nil {
		return nil, errorx.StatusInternalErrorF("begin login err=%v", err)
	}

	sid, _ := srv.store.GenSessionId()
	srv.store.SaveSession(sid, session)

	return &iampb.PasskeyLoginChallengeResponse{
		SessionId: sid,
		Options:   protox.Struct(options),
	}, nil
}

// PasskeyRegisterChallenge starts signing up a new user with a passkey.
// An email that already belongs to a user is rejected: that user adds a
// passkey with PasskeyLinkChallenge, once signed in.
func (srv *IAMService) PasskeyRegisterChallenge(ctx context.Context,
	req *iampb.PasskeyRegisterChallengeRequest,
) (*iampb.PasskeyRegisterChallengeResponse, error) {

	email := req.GetEmailOrUsername()
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errorx.StatusInvalidDataF(
			"passkey registration needs an email, got %s", email)
	}

	if err := srv.emailMustFree(ctx, email); err != nil {
		return nil, err
	}

	options, sid, err := srv.beginRegistration(&passkey.Registration{
		Email:  email,
		Handle: rand.NewId(""),
	})
	if err != nil {
		return nil, err
	}

	return &iampb.PasskeyRegisterChallengeResponse{
		Options:   options,
		SessionId: sid,
	}, nil
}

// PasskeyLinkChallenge starts adding a passkey to the signed-in user.
func (srv *IAMService) PasskeyLinkChallenge(ctx context.Context,
	_ *iampb.PasskeyLinkChallengeRequest,
) (*iampb.PasskeyLinkChallengeResponse, error) {

	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return nil, err
	}

	user, err := srv.users.Get(ctx, auth.GetUserId())
	if err != nil {
		return nil, err
	}

	reg, err := srv.linkRegistration(ctx, user)
	if err != nil {
		return nil, err
	}

	options, sid, err := srv.beginRegistration(reg)
	if err != nil {
		return nil, err
	}

	return &iampb.PasskeyLinkChallengeResponse{
		Options:   options,
		SessionId: sid,
	}, nil
}

// linkRegistration returns the registration adding a passkey to user.
func (srv *IAMService) linkRegistration(ctx context.Context,
	user *iampb.User) (*passkey.Registration, error) {

	acc, err := srv.findPasskeyAccount(ctx, user.GetId())
	if err != nil {
		return nil, err
	}

	reg := &passkey.Registration{
		UserId: user.GetId(),
		Email:  user.GetEmail(),
		Handle: rand.NewId(""),
	}
	if acc != nil {
		// every passkey of a user shares the handle of its account
		reg.Handle = acc.GetProviderUserId()
	}

	return reg, nil
}

// beginRegistration returns the WebAuthn creation options and the id of
// the session that PasskeyRegisterVerify finishes.
func (srv *IAMService) beginRegistration(
	reg *passkey.Registration) (*structpb.Struct, string, error) {

	opt, ss, err := srv.auth.BeginRegistration(passkeyAccount(reg))
	if err != nil {
		return nil, "",
			errorx.StatusInternalErrorF("can't begin registration: %v", err)
	}

	sid, err := srv.store.GenSessionId()
	if err != nil {
		return nil, "",
			errorx.StatusInternalErrorF("can't generate session id: %v", err)
	}

	srv.store.SaveSession(sid, ss)
	srv.store.SaveRegistration(sid, reg)

	return protox.Struct(opt), sid, nil
}

// PasskeyRegisterVerify finishes a registration started by
// PasskeyRegisterChallenge or PasskeyLinkChallenge. The user and account
// are written only after the authenticator response is verified.
func (srv *IAMService) PasskeyRegisterVerify(ctx context.Context,
	req *iampb.PasskeyRegisterVerifyRequest,
) (*iampb.PasskeyRegisterVerifyResponse, error) {

	sid := req.GetSessionId()
	ss, ok := srv.store.GetSession(sid)
	if !ok {
		return nil, errorx.StatusNotFoundF("session not found=%s", sid)
	}

	reg, ok := srv.store.TakeRegistration(sid)
	if !ok {
		return nil, errorx.StatusNotFoundF("session not found=%s", sid)
	}

	srv.store.DeleteSession(sid)

	if err := authorizeLink(ctx, reg); err != nil {
		return nil, err
	}

	credential, err := srv.createCredential(
		reg, ss, req.GetCredentialCreationResponse())
	if err != nil {
		return nil, err
	}

	if err = srv.savePasskey(ctx, reg, credential); err != nil {
		return nil, err
	}

	return &iampb.PasskeyRegisterVerifyResponse{}, nil
}

// authorizeLink checks that a registration adding a passkey to an
// existing user is finished by that user.
func authorizeLink(ctx context.Context, reg *passkey.Registration) error {
	if reg.UserId == "" {
		return nil
	}

	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return err
	}

	if auth.GetUserId() != reg.UserId {
		return errorx.StatusForbiddenF("passkey belongs to another user")
	}

	return nil
}

func (srv *IAMService) createCredential(reg *passkey.Registration,
	ss *webauthn.SessionData, data []byte) (*webauthn.Credential, error) {

	var ccr protocol.CredentialCreationResponse
	if err := json.Unmarshal(data, &ccr); err != nil {
		return nil, err
	}

	parsedCCR, err := ccr.Parse()
	if err != nil {
		return nil, err
	}

	credential, err := srv.auth.CreateCredential(
		passkeyAccount(reg), *ss, parsedCCR)
	if err != nil {
		return nil,
			errorx.StatusInternalErrorF("can't finish registration: %v", err)
	}

	return credential, nil
}

// savePasskey stores a verified credential: on a new user when reg signs
// up, otherwise on the passkey account of reg.UserId.
func (srv *IAMService) savePasskey(ctx context.Context,
	reg *passkey.Registration, credential *webauthn.Credential) error {

	acc := passkeyAccount(reg)
	acc.AddCredential(credential)

	if reg.UserId == "" {
		// the email may have been taken since the challenge
		if err := srv.emailMustFree(ctx, reg.Email); err != nil {
			return err
		}

		_, err := srv.createUserAccount(ctx,
			&iampb.User{FullName: reg.Email, Email: reg.Email}, acc)
		return err
	}

	existing, err := srv.findPasskeyAccount(ctx, reg.UserId)
	if err != nil {
		return err
	}

	if existing == nil {
		_, err = srv.accounts.Create(ctx, acc)
		return err
	}

	return srv.addCredential(ctx, existing, reg, credential)
}

func (srv *IAMService) addCredential(ctx context.Context,
	acc *accrepos.AccountX, reg *passkey.Registration,
	credential *webauthn.Credential) error {
	// the authenticator stored reg.Handle; under another handle the
	// credential could never sign in
	if acc.GetProviderUserId() != reg.Handle {
		return errorx.StatusInvalidDataF(
			"passkey registration is stale, start again")
	}

	acc.AddCredential(credential)

	return srv.accounts.Update(ctx, acc)
}

// findPasskeyAccount returns the passkey account of userId, or nil if the
// user has none.
func (srv *IAMService) findPasskeyAccount(ctx context.Context,
	userId string) (*accrepos.AccountX, error) {
	return found(srv.accounts.GetByUserId(
		ctx, iampb.Provider_PROVIDER_PASSKEY, userId))
}

// passkeyAccount builds the account a registration is verified against
// and, once verified, saved as.
func passkeyAccount(reg *passkey.Registration) *accrepos.AccountX {
	return &accrepos.AccountX{Account: &iampb.Account{
		UserId:         reg.UserId,
		Email:          reg.Email,
		Provider:       iampb.Provider_PROVIDER_PASSKEY,
		ProviderUserId: reg.Handle,
	}}
}
