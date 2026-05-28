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

package iamsvc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"

	"github.com/sentinez/core/storage/dbx/postgres"
	accrepos "github.com/sentinez/modules/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/modules/iam/v1/repos/users"
	"github.com/sentinez/modules/pkg/crypto"
	"github.com/sentinez/modules/pkg/passkey"
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/iam/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/errorx"
	"github.com/sentinez/shared/protobuf/protox"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ iampb.IdentityAccessManagementServiceServer = (*IAMService)(nil)

func New(config *confpb.Config,
	tx *postgres.Tx,
	store passkey.Store,
	users usersrepo.IUser,
	account accrepos.IAccount,
) *IAMService {

	wauth := passkey.NewWebAuthn(config)

	return &IAMService{
		config:   config,
		tx:       tx,
		users:    users,
		accounts: account,
		store:    store,
		auth:     wauth,
	}
}

type IAMService struct {
	config   *confpb.Config
	tx       *postgres.Tx
	users    usersrepo.IUser
	accounts accrepos.IAccount
	store    passkey.Store
	auth     *webauthn.WebAuthn
}

// PasskeyLoginVerify implements iampb.IdentityAccessManagementServiceServer.
// nolint:funlen
func (srv *IAMService) PasskeyLoginVerify(ctx context.Context,
	req *iampb.PasskeyLoginVerifyRequest,
) (*iampb.PasskeyLoginVerifyResponse, error) {

	sid := req.GetSessionId()
	ss, ok := srv.store.GetSession(sid)
	if !ok {
		return nil, errorx.StatusNotFoundF("session not found id=%s", sid)
	}

	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, string(ss.UserID))
	if err != nil {
		return nil, err
	}

	var car protocol.CredentialAssertionResponse
	err = json.Unmarshal(req.GetCredentialAssertionData(), &car)
	if err != nil {
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

	acc.UpdateCredential(credential)
	if err = srv.accounts.Update(ctx, acc); err != nil {
		return nil, err
	}

	srv.store.DeleteSession(sid)

	sid, _ = srv.store.GenSessionID()
	srv.store.SaveSession(sid, &webauthn.SessionData{
		Expires: time.Now().Add(time.Hour * 2),
	})

	user, err := srv.users.Get(ctx, acc.GetUserId())
	if err != nil {
		zlog.Debugf("faild to get user by username or email")
		return nil, err
	}

	console := typepb.Console_CONSOLE_PORTAL
	if acc.GetUsername() == "admin" {
		console = typepb.Console_CONSOLE_ADMIN
	}

	accessToken, err := crypto.TokenGenerator(&typepb.Context{
		Name:     user.GetFullName(),
		ExpireAt: timestamppb.New(time.Now().Add(time.Hour)),
		UserId:   user.GetId(),
		Console:  console,
	})
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

	zlog.Infof("PasskeyLoginChallenge req = %v", req)

	emailOrUsername := req.GetEmailOrUsername()
	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, emailOrUsername)
	if err != nil {
		if !errorx.NotRowsNotFound(err) {
			return nil, errorx.StatusNotFoundF(
				"not found username or email=%s", emailOrUsername)
		}
		return nil, err
	}

	options, session, err := srv.auth.BeginLogin(acc)
	if err != nil {
		return nil, errorx.StatusInternalErrorF("begin login err=%v", err)
	}

	sid, _ := srv.store.GenSessionID()
	srv.store.SaveSession(sid, session)

	opts := protox.Struct(options)

	return &iampb.PasskeyLoginChallengeResponse{
		SessionId: sid,
		Options:   opts,
	}, nil
}

// PasskeyRegisterVerify implements iampb.IdentityAccessManagementServiceServer.
// nolint:funlen
func (srv *IAMService) PasskeyRegisterVerify(ctx context.Context,
	req *iampb.PasskeyRegisterVerifyRequest,
) (*iampb.PasskeyRegisterVerifyResponse, error) {

	ssId := req.GetSessionId()

	zlog.Debugf("[iampb][service][PasskeyRegisterVerify] get session=%s", ssId)
	ss, ok := srv.store.GetSession(ssId)
	if !ok {
		zlog.Debug("iampb: session not found")
		return nil, errorx.StatusNotFoundF("session not found=%s", ssId)
	}

	// ss.UserId is email, return by AccountX.WebAuthnID(), account_x.go
	acc, err := srv.store.GetAndDeleteAccount(string(ss.UserID))
	if err != nil {
		zlog.Debugf("iampb: get %s account err: %v", string(ss.UserID), err)
		return nil, err
	}

	accX, err := srv.createAccountExtend(ctx, acc)
	if err != nil {
		zlog.Debugf("iampb: create account extend: %v", err)
		return nil, err
	}

	var ccr protocol.CredentialCreationResponse
	err = json.Unmarshal(req.GetCredentialCreationResponse(), &ccr)
	if err != nil {
		return nil, err
	}

	parsedCCR, err := ccr.Parse()
	if err != nil {
		return nil, err
	}

	credential, err := srv.auth.CreateCredential(accX, *ss, parsedCCR)
	if err != nil {
		return nil,
			errorx.StatusInternalErrorF("can't finish registration: %v", err)
	}

	accX.AddCredential(credential)
	if err = srv.accounts.Update(ctx, accX); err != nil {
		return nil, err
	}

	srv.store.DeleteSession(ssId)

	return &iampb.PasskeyRegisterVerifyResponse{}, nil
}

func (srv *IAMService) PasskeyRegisterChallenge(_ context.Context,
	req *iampb.PasskeyRegisterChallengeRequest,
) (*iampb.PasskeyRegisterChallengeResponse, error) {

	acc, err := srv.store.GetOrCreateAccount(req.GetEmailOrUsername())
	if err != nil {
		return nil, err
	}

	opt, ss, err := srv.auth.BeginRegistration(&accrepos.AccountX{Account: acc})
	if err != nil {
		return nil,
			errorx.StatusInternalErrorF("can't begin registration: %v", err)
	}

	t, err := srv.store.GenSessionID()
	if err != nil {
		return nil,
			errorx.StatusInternalErrorF("can't generate session id: %v", err)
	}

	zlog.Debugf("[iampb][service][PasskeyRegisterChallenge] save session=%s", t)
	srv.store.SaveSession(t, ss)

	options := protox.Struct(opt)

	return &iampb.PasskeyRegisterChallengeResponse{
		Options:   options,
		SessionId: t,
	}, nil
}

func (srv *IAMService) createAccountExtend(
	ctx context.Context, account *iampb.Account) (*accrepos.AccountX, error) {
	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, account.GetEmail())
	if err != nil && errorx.NotRowsNotFound(err) {
		zlog.Errorf("GetByUsernameOrEmail err=%v", err)
		return nil, err
	}

	if acc.GetId() == "" {
		createReq := &iampb.CreateAccountRequest{
			Email:    account.GetEmail(),
			Username: account.GetUsername(),
		}

		createReq.Password, _ = rand.RandomString(5)

		createResp, err := srv.CreateAccount(ctx, createReq)
		if err != nil {
			zlog.Errorf("CreateAccount err=%v", err)
			return nil, err
		}

		acc, err = srv.accounts.Get(ctx, createResp.GetAccountId())
		if err != nil {
			zlog.Errorf("accounts.Get err=%v", err)
			return nil, err
		}
	}

	return acc, nil
}

func (srv *IAMService) Config() *confpb.EnvConfig {
	return srv.config.GetEnv()
}

func (srv *IAMService) ListAccounts(ctx context.Context,
	request *iampb.ListAccountsRequest) (*iampb.ListAccountsResponse, error) {

	resp, err := srv.accounts.List(ctx, request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (srv *IAMService) Status(ctx context.Context,
	request *iampb.StatusRequest) (*iampb.StatusResponse, error) {

	_ = ctx
	_ = request

	return &iampb.StatusResponse{Msg: "OK"}, nil
}

func (srv *IAMService) UsernameOrEmailMustUnique(ctx context.Context,
	username, email string) error {

	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, username)
	if errorx.NotRowsNotFound(err) {
		return err
	}

	if acc.GetId() != "" {
		return errorx.StatusAlreadyExistsF(
			"username %s already exists", acc.GetUsername())
	}

	acc, err = srv.accounts.GetByUsernameOrEmail(ctx, email)
	if errorx.NotRowsNotFound(err) {
		return err
	}
	if acc.GetId() != "" {
		return errorx.StatusAlreadyExistsF(
			"email %s already exists", acc.GetEmail())
	}

	return nil
}

func (srv *IAMService) CreateAccount(ctx context.Context,
	request *iampb.CreateAccountRequest) (*iampb.CreateAccountResponse, error) {

	if err := srv.UsernameOrEmailMustUnique(
		ctx, request.GetUsername(), request.GetEmail()); err != nil {
		return nil, err
	}

	txss, err := srv.tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	accID, err := srv.createAccountWithTX(ctx, txss, request)
	if err != nil {
		return nil, err
	}

	return &iampb.CreateAccountResponse{AccountId: accID}, nil
}

func (srv *IAMService) createAccountWithTX(ctx context.Context,
	txss *postgres.TxSession, req *iampb.CreateAccountRequest) (string, error) {
	pw, err := crypto.HashPassword(req.GetPassword())
	if err != nil {
		return "", err
	}

	user, err := srv.users.WithTX(txss).Create(ctx, &iampb.User{
		FullName:    req.GetFullName(),
		EmailBackup: req.GetEmail(),
		PhoneNumber: req.GetPhoneNumber(),
	})
	if err != nil {
		_ = txss.Rollback(ctx)
		return "", err
	}

	acc, err := srv.accounts.WithTX(txss).Create(ctx, &accrepos.AccountX{
		Account: &iampb.Account{
			UserId:       user.GetId(),
			Email:        req.GetEmail(),
			Username:     req.GetUsername(),
			PasswordHash: pw,
		},
	})
	if err != nil {
		_ = txss.Rollback(ctx)
		return "", err
	}

	_ = txss.Commit(ctx)
	return acc.GetId(), nil
}

func (srv *IAMService) GetAccountByUsernameOrEmail(
	ctx context.Context, usernameOrEmail string) (*accrepos.AccountX, error) {

	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, usernameOrEmail)
	if err != nil {
		if errorx.Is(err, pgx.ErrNoRows) {
			return &accrepos.AccountX{}, nil
		}

		return nil, err
	}

	return acc, nil
}

func (srv *IAMService) Login(ctx context.Context,
	req *iampb.LoginRequest) (*iampb.LoginResponse, error) {

	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, req.GetEmailOrUsername())
	if err != nil {
		zlog.Debugf("faild to get account by username or email")
		return nil, err
	}

	if !crypto.CheckPasswordHash(req.GetPassword(), acc.GetPasswordHash()) {
		return nil,
			errorx.StatusUnauthorizedF("username, email or password is wrong!")
	}

	user, err := srv.users.Get(ctx, acc.GetUserId())
	if err != nil {
		zlog.Debugf("faild to get user by username or email")
		return nil, err
	}

	console := typepb.Console_CONSOLE_PORTAL
	if acc.GetUsername() == "admin" {
		console = typepb.Console_CONSOLE_ADMIN
	}

	accessToken, err := crypto.TokenGenerator(&typepb.Context{
		Name:     user.GetFullName(),
		ExpireAt: timestamppb.New(time.Now().Add(time.Hour)),
		UserId:   user.GetId(),
		Console:  console,
	})
	if err != nil {
		return nil, err
	}

	return &iampb.LoginResponse{User: user, AccessToken: accessToken}, nil
}

func (srv *IAMService) CreateUser(ctx context.Context,
	request *iampb.CreateUserRequest) (*iampb.CreateUserResponse, error) {

	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, request.GetEmail())
	if errorx.NotRowsNotFound(err) {
		return nil, err
	}

	if acc.GetId() != "" {
		return nil,
			errorx.StatusAlreadyExistsF(
				"email %s already exists", request.GetEmail())
	}

	user, err := srv.users.Create(ctx, &iampb.User{
		FullName:    request.GetFullName(),
		EmailBackup: request.GetEmail(),
		PhoneNumber: request.GetPhoneNumber(),
	})
	if err != nil {
		return nil, err
	}

	return &iampb.CreateUserResponse{UserId: user.Id}, nil
}

func (srv *IAMService) GetUser(ctx context.Context,
	request *iampb.GetUserRequest) (*iampb.GetUserResponse, error) {
	if request.GetId() != "" {
		user, err := srv.users.Get(ctx, request.GetId())
		if err != nil {
			return nil, err
		}
		return &iampb.GetUserResponse{User: user}, nil
	}

	if request.GetEmail() != "" {
		acc, err := srv.accounts.GetByUsernameOrEmail(ctx, request.GetEmail())
		if err != nil {
			return nil, err
		}

		user, err := srv.users.Get(ctx, acc.GetUserId())
		if err != nil {
			return nil, err
		}

		return &iampb.GetUserResponse{User: user}, nil
	}

	return nil, errorx.StatusInvalidDataF(
		"invalid argument: must provide either id or username")
}

func (srv *IAMService) ListUsers(ctx context.Context,
	request *iampb.ListUsersRequest) (*iampb.ListUsersResponse, error) {

	users, err := srv.users.List(ctx, request)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (srv *IAMService) DeleteUser(ctx context.Context,
	request *iampb.DeleteUserRequest) (*iampb.DeleteUserResponse, error) {

	if err := srv.users.Delete(ctx, request.GetId()); err != nil {
		return nil, err
	}

	return &iampb.DeleteUserResponse{}, nil
}

func (srv *IAMService) UpdateUser(ctx context.Context,
	request *iampb.UpdateUserRequest) (*iampb.UpdateUserResponse, error) {

	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, request.GetEmail())
	if errorx.NotRowsNotFound(err) {
		return nil, err
	}

	if acc.GetId() == "" {
		return nil, errorx.StatusNotFoundF(
			"email %s not found", request.GetEmail())
	}

	user, err := srv.users.Get(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	copyUserUpdateParams(user, request)

	err = srv.users.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return &iampb.UpdateUserResponse{}, nil
}

func copyUserUpdateParams(dest *iampb.User, req *iampb.UpdateUserRequest) {

	if req.GetEmail() != "" {
		dest.EmailBackup = req.GetEmail()
	}

	if req.GetFullName() != "" {
		dest.FullName = req.GetFullName()
	}

	if req.GetPhoneNumber() != "" {
		dest.PhoneNumber = req.GetPhoneNumber()
	}
}
