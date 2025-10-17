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

package iamsvc

import (
	"context"
	"encoding/json"
	"net/mail"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	modelpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/model/v1"
	accrepos "github.com/sentinez/sentinez/internal/core/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"
	"github.com/sentinez/sentinez/pkg/security/passkey"
	"github.com/sentinez/sentinez/pkg/security/perms"
	"github.com/sentinez/sentinez/pkg/storage/database/postgres"
	"github.com/sentinez/sentinez/pkg/x/cryptox"
	"github.com/sentinez/sentinez/pkg/x/errorx"
	"github.com/sentinez/sentinez/pkg/x/protobuf/protox"
	"github.com/sentinez/sentinez/pkg/x/randx"
	"github.com/sentinez/sentinez/pkg/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ iam.IdentityAccessManagementServiceServer = (*IAMService)(nil)

func New(config *common.AppConfig,
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
	config   *common.AppConfig
	tx       *postgres.Tx
	users    usersrepo.IUser
	accounts accrepos.IAccount
	store    passkey.Store
	auth     *webauthn.WebAuthn
}

// PasskeyLoginFinish implements iam.IdentityAccessManagementServiceServer.
// nolint:funlen
func (srv *IAMService) PasskeyLoginFinish(ctx context.Context,
	req *iam.PasskeyLoginFinishRequest,
) (*iam.PasskeyLoginFinishResponse, error) {

	sid := req.GetSessionId()
	ss, ok := srv.store.GetSession(sid)
	if !ok {
		return nil, errorx.StatusNotFoundF("session not found id=%s", sid)
	}

	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, string(ss.UserID))
	if err != nil {
		return nil, err
	}

	if acc.GetId() == "" {
		acc = &accrepos.AccountX{}
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
	if _, err := srv.accounts.Update(ctx, acc); err != nil {
		return nil, err
	}

	srv.store.DeleteSession(sid)

	sid, _ = srv.store.GenSessionID()
	srv.store.SaveSession(sid, &webauthn.SessionData{
		Expires: time.Now().Add(time.Hour * 2),
	})

	return &iam.PasskeyLoginFinishResponse{}, nil
}

// PasskeyLoginStart implements iam.IdentityAccessManagementServiceServer.
func (srv *IAMService) PasskeyLoginStart(
	ctx context.Context,
	req *iam.PasskeyLoginStartRequest,
) (*iam.PasskeyLoginStartResponse, error) {

	emailOrUsername := req.GetEmailOrUsername()
	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, emailOrUsername)
	if err != nil {
		return nil, err
	}

	options, session, err := srv.auth.BeginLogin(acc)
	if err != nil {
		return nil, errorx.StatusInternalErrorF("begin login err=%v", err)
	}

	sid, _ := srv.store.GenSessionID()
	srv.store.SaveSession(sid, session)

	opts := protox.Struct(options)

	return &iam.PasskeyLoginStartResponse{
		SessionId: sid,
		Options:   opts,
	}, nil
}

// PasskeyRegisterFinish implements iam.IdentityAccessManagementServiceServer.
func (srv *IAMService) PasskeyRegisterFinish(ctx context.Context,
	req *iam.PasskeyRegisterFinishRequest,
) (*iam.PasskeyRegisterFinishResponse, error) {

	ssId := req.GetSessionId()

	zlog.Debugf("[iam][service][PasskeyRegisterFinish] get session=%s", ssId)
	ss, ok := srv.store.GetSession(ssId)
	if !ok {
		return nil, errorx.StatusNotFoundF("session not found=%s", ssId)
	}

	acc, err := srv.accounts.Get(ctx, string(ss.UserID))
	if err != nil {
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

	credential, err := srv.auth.CreateCredential(acc, *ss, parsedCCR)
	if err != nil {
		return nil,
			errorx.StatusInternalErrorF("can't finish registration: %v", err)
	}

	acc.AddCredential(credential)
	if _, err = srv.accounts.Update(ctx, acc); err != nil {
		return nil, err
	}

	srv.store.DeleteSession(ssId)

	return &iam.PasskeyRegisterFinishResponse{}, nil
}

// PasskeyRegisterStart implements iam.IdentityAccessManagementServiceServer.
func (srv *IAMService) PasskeyRegisterStart(ctx context.Context,
	req *iam.PasskeyRegisterStartRequest,
) (*iam.PasskeyRegisterStartResponse, error) {

	acc, err := srv.getOrCreateAccount(ctx, req.GetEmailOrUsername())
	if err != nil {
		return nil, err
	}

	opt, ss, err := srv.auth.BeginRegistration(acc)
	if err != nil {
		return nil,
			errorx.StatusInternalErrorF("can't begin registration: %v", err)
	}

	t, err := srv.store.GenSessionID()
	if err != nil {
		return nil,
			errorx.StatusInternalErrorF("can't generate session id: %v", err)
	}

	zlog.Debugf("[iam][service][PasskeyRegisterStart] save session=%s", t)
	srv.store.SaveSession(t, ss)

	options := protox.Struct(opt)

	return &iam.PasskeyRegisterStartResponse{
		Options:   options,
		SessionId: t,
	}, nil
}

func (srv *IAMService) getOrCreateAccount(
	ctx context.Context, usernameOrEmail string) (*accrepos.AccountX, error) {
	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, usernameOrEmail)
	if err != nil && errorx.NotRowsNotFound(err) {
		zlog.Errorf("GetByUsernameOrEmail err=%v", err)
		return nil, err
	}

	if acc.GetId() == "" {
		createReq := &iam.CreateAccountRequest{}
		_, err = mail.ParseAddress(usernameOrEmail)
		if err != nil {
			createReq.Username = usernameOrEmail
		}

		if err == nil {
			createReq.Email = usernameOrEmail
		}

		createReq.Password, _ = randx.RandomString(5)

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

func (srv *IAMService) Config() *common.EnvConfig {
	return srv.config.GetEnvConf()
}

func (srv *IAMService) ListAccounts(ctx context.Context,
	request *iam.ListAccountsRequest) (*iam.ListAccountsResponse, error) {

	resp, err := srv.accounts.List(ctx, request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (srv *IAMService) Status(ctx context.Context,
	request *iam.StatusRequest) (*iam.StatusResponse, error) {

	_ = ctx
	_ = request

	return &iam.StatusResponse{Msg: "OK"}, nil
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
	request *iam.CreateAccountRequest) (*iam.CreateAccountResponse, error) {

	if err := srv.UsernameOrEmailMustUnique(
		ctx, request.GetUsername(), request.GetEmail()); err != nil {
		return nil, err
	}

	txss, err := srv.tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	accID, err := srv.createAccount(ctx, txss, request)
	if err != nil {
		return nil, err
	}

	return &iam.CreateAccountResponse{AccountId: accID}, nil
}

func (srv *IAMService) createAccount(ctx context.Context,
	txss *postgres.TxSession, req *iam.CreateAccountRequest) (string, error) {

	pw, err := cryptox.HashPassword(req.GetPassword())
	if err != nil {
		return "", err
	}

	user, err := srv.users.WithTX(txss).Create(ctx, &iam.User{
		Metadata: &modelpb.Metadata{
			CreatedBy: req.GetUsername(),
			UpdatedBy: req.GetUsername(),
		},
		FullName:    req.GetFullName(),
		Email:       req.GetEmail(),
		PhoneNumber: req.GetPhoneNumber(),
	})
	if err != nil {
		_ = txss.Rollback(ctx)
		return "", err
	}

	acc, err := srv.accounts.WithTX(txss).Create(ctx, &accrepos.AccountX{
		Account: &iam.Account{
			UserId:   user.GetId(),
			Email:    req.GetEmail(),
			Username: req.GetUsername(),
			Password: pw,
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
	req *iam.LoginRequest) (*iam.LoginResponse, error) {

	acc, err := srv.accounts.GetByUsernameOrEmail(ctx, req.GetEmailOrUsername())
	if err != nil {
		zlog.Debugf("faild to get account by username or email")
		return nil, err
	}

	if !cryptox.CheckPasswordHash(req.GetPassword(), acc.GetPassword()) {
		return nil,
			errorx.StatusUnauthorizedF("username, email or password is wrong!")
	}

	user, err := srv.users.Get(ctx, acc.GetUserId())
	if err != nil {
		zlog.Debugf("faild to get user by username or email")
		return nil, err
	}
	perm := perms.DefaultOwner()
	if acc.GetUsername() == "admin" {
		perm = perms.Add(perm, common.Permission_PERMISSION_ROOT)
	}
	accessToken, err := cryptox.TokenGenerator(srv.config.GetEnvConf(),
		&common.Context{
			Name:              user.GetFullName(),
			ExpireAt:          timestamppb.New(time.Now().Add(time.Hour)),
			UserId:            user.GetId(),
			PermissionBitwise: perm,
		},
	)
	if err != nil {
		return nil, err
	}

	return &iam.LoginResponse{User: user, AccessToken: accessToken}, nil
}

func (srv *IAMService) CreateUser(ctx context.Context,
	request *iam.CreateUserRequest) (*iam.CreateUserResponse, error) {

	user, err := srv.users.GetByFullnameOrEmail(ctx, request.GetEmail())
	if errorx.NotRowsNotFound(err) {
		return nil, err
	}

	if user.GetId() != "" {
		return nil,
			errorx.StatusAlreadyExistsF(
				"email %s already exists", request.GetEmail())
	}

	user, err = srv.users.Create(ctx, &iam.User{
		FullName:    request.GetFullName(),
		Email:       request.GetEmail(),
		PhoneNumber: request.GetPhoneNumber(),
	})
	if err != nil {
		return nil, err
	}

	return &iam.CreateUserResponse{UserId: user.Id}, nil
}

func (srv *IAMService) GetUser(ctx context.Context,
	request *iam.GetUserRequest) (*iam.GetUserResponse, error) {
	if request.GetId() != "" {
		user, err := srv.users.Get(ctx, request.GetId())
		if err != nil {
			return nil, err
		}
		return &iam.GetUserResponse{User: user}, nil
	}

	if request.GetEmail() != "" {
		user, err := srv.users.GetByFullnameOrEmail(ctx, request.GetEmail())
		if err != nil {
			return nil, err
		}

		return &iam.GetUserResponse{User: user}, nil
	}

	return nil, errorx.StatusInvalidDataF(
		"invalid argument: must provide either id or username")
}

func (srv *IAMService) ListUsers(ctx context.Context,
	request *iam.ListUsersRequest) (*iam.ListUsersResponse, error) {

	users, err := srv.users.List(ctx, request)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (srv *IAMService) DeleteUser(ctx context.Context,
	request *iam.DeleteUserRequest) (*iam.DeleteUserResponse, error) {

	if err := srv.users.Delete(ctx, request.GetId()); err != nil {
		return nil, err
	}

	return &iam.DeleteUserResponse{}, nil
}

func (srv *IAMService) UpdateUser(ctx context.Context,
	request *iam.UpdateUserRequest) (*iam.UpdateUserResponse, error) {

	user, err := srv.users.GetByFullnameOrEmail(ctx, request.GetEmail())
	if errorx.NotRowsNotFound(err) {
		return nil, err
	}

	if user.GetId() != "" {
		return nil, errorx.StatusAlreadyExistsF(
			"email %s already exists", request.GetEmail())
	}

	user, err = srv.users.Get(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	copyUserUpdateParams(user, request)

	_, err = srv.users.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return &iam.UpdateUserResponse{}, nil
}

func copyUserUpdateParams(dest *iam.User, req *iam.UpdateUserRequest) {

	if req.GetEmail() != "" {
		dest.Email = req.GetEmail()
	}

	if req.GetEmail() != "" {
		dest.FullName = req.GetFullName()
	}

	if req.GetPhoneNumber() != "" {
		dest.PhoneNumber = req.GetPhoneNumber()
	}
}
