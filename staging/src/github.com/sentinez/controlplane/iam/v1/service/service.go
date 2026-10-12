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
	"errors"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	accrepos "github.com/sentinez/controlplane/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/controlplane/iam/v1/repos/users"
	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/controlplane/pkg/passkey"
	"github.com/sentinez/core/storage/dbx/postgres"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/shared/errorx"
	"github.com/sentinez/shared/zlog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ iampb.IdentityAccessManagementServiceServer = (*IAMService)(nil)

func New(config *settingpb.Config,
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
	config   *settingpb.Config
	tx       *postgres.Tx
	users    usersrepo.IUser
	accounts accrepos.IAccount
	store    passkey.Store
	auth     *webauthn.WebAuthn
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

// UsernameOrEmailMustUnique reports an AlreadyExists error if a local
// account already uses username or email. Accounts of other providers may
// share the email: they are other login methods of the same user.
func (srv *IAMService) UsernameOrEmailMustUnique(ctx context.Context,
	username, email string) error {
	if username != "" {
		if err := srv.usernameMustFree(ctx, username); err != nil {
			return err
		}
	}

	acc, err := srv.findLocalAccount(ctx, email)
	if err != nil {
		return err
	}

	if acc != nil {
		return errorx.StatusAlreadyExistsF("email %s already exists", email)
	}

	return nil
}

func (srv *IAMService) usernameMustFree(
	ctx context.Context, username string) error {
	// the admin signs in with this name and has no account row
	if username == srv.config.Get(settingpb.Senz_SENZ_ADMIN_USERNAME) {
		return errorx.StatusAlreadyExistsF(
			"username %s already exists", username)
	}

	acc, err := srv.findLocalAccount(ctx, username)
	if err != nil {
		return err
	}

	if acc != nil {
		return errorx.StatusAlreadyExistsF(
			"username %s already exists", username)
	}

	return nil
}

// findAccount returns the account of any provider matching
// usernameOrEmail, or nil if there is none. Any other lookup failure is
// returned as an error.
func (srv *IAMService) findAccount(ctx context.Context,
	usernameOrEmail string) (*accrepos.AccountX, error) {
	return found(srv.accounts.GetByUsernameOrEmail(ctx, usernameOrEmail))
}

// findLocalAccount is findAccount limited to the local provider.
func (srv *IAMService) findLocalAccount(ctx context.Context,
	usernameOrEmail string) (*accrepos.AccountX, error) {
	return found(srv.accounts.GetByProvider(
		ctx, iampb.Provider_PROVIDER_LOCAL, usernameOrEmail))
}

// findUser returns the user owning email, or nil if there is none.
func (srv *IAMService) findUser(
	ctx context.Context, email string) (*iampb.User, error) {
	return found(srv.users.GetByEmail(ctx, email))
}

// found turns errorx.ErrNotFound into a nil result.
func found[T any](v *T, err error) (*T, error) {
	if errors.Is(err, errorx.ErrNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return v, nil
}

// CreateAccount signs up a new user with a local (password) account. An
// email that already belongs to a user is rejected: a login method is
// only added to an existing user by that user, once signed in.
func (srv *IAMService) CreateAccount(ctx context.Context,
	request *iampb.CreateAccountRequest) (*iampb.CreateAccountResponse, error) {

	if err := srv.UsernameOrEmailMustUnique(
		ctx, request.GetUsername(), request.GetEmail()); err != nil {
		return nil, err
	}

	if err := srv.emailMustFree(ctx, request.GetEmail()); err != nil {
		return nil, err
	}

	pw, err := crypto.HashPassword(request.GetPassword())
	if err != nil {
		return nil, err
	}

	acc, err := srv.createUserAccount(ctx, &iampb.User{
		FullName: request.GetFullName(),
		Email:    request.GetEmail(),
	}, &accrepos.AccountX{Account: &iampb.Account{
		Email:        request.GetEmail(),
		Username:     request.GetUsername(),
		PasswordHash: pw,
		Provider:     iampb.Provider_PROVIDER_LOCAL,
	}})
	if err != nil {
		return nil, err
	}

	return &iampb.CreateAccountResponse{AccountId: acc.GetId()}, nil
}

// emailMustFree reports an AlreadyExists error if a user owns email.
func (srv *IAMService) emailMustFree(ctx context.Context, email string) error {
	user, err := srv.findUser(ctx, email)
	if err != nil {
		return err
	}

	if user != nil {
		return errorx.StatusAlreadyExistsF("email %s already exists", email)
	}

	return nil
}

// createUserAccount creates user and its first account in one transaction.
func (srv *IAMService) createUserAccount(ctx context.Context,
	user *iampb.User, acc *accrepos.AccountX) (*accrepos.AccountX, error) {
	txss, err := srv.tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	user.ControlPlane = typepb.ControlPlane_CONTROL_PLANE_PORTAL
	user, err = srv.users.WithTX(txss).Create(ctx, user)
	if err != nil {
		_ = txss.Rollback(ctx)
		return nil, err
	}

	acc.UserId = user.GetId()
	acc, err = srv.accounts.WithTX(txss).Create(ctx, acc)
	if err != nil {
		_ = txss.Rollback(ctx)
		return nil, err
	}

	if err = txss.Commit(ctx); err != nil {
		return nil, err
	}

	return acc, nil
}

// newAccessToken issues a token for user on the user's own control plane.
func newAccessToken(user *iampb.User) (string, error) {
	controlPlane := user.GetControlPlane()
	if controlPlane == typepb.ControlPlane_CONTROL_PLANE_UNSPECIFIED {
		controlPlane = typepb.ControlPlane_CONTROL_PLANE_PORTAL
	}

	return crypto.TokenGenerator(&typepb.Context{
		Name:         user.GetFullName(),
		ExpireAt:     timestamppb.New(time.Now().Add(time.Hour)),
		UserId:       user.GetId(),
		ControlPlane: controlPlane,
	})
}

func (srv *IAMService) loginAdmin(
	req *iampb.LoginRequest) (*iampb.LoginResponse, error) {

	if req.GetPassword() != srv.config.Get(settingpb.Senz_SENZ_ADMIN_PASSWORD) {
		return nil, errorx.StatusUnauthorizedF(
			"username, email or password is wrong!")
	}

	accessToken, err := crypto.TokenGenerator(&typepb.Context{
		Name:         req.GetEmailOrUsername(),
		ExpireAt:     timestamppb.New(time.Now().Add(time.Hour)),
		UserId:       "sentinez.admin",
		ControlPlane: typepb.ControlPlane_CONTROL_PLANE_ADMIN,
	})
	if err != nil {
		return nil, err
	}

	return &iampb.LoginResponse{User: &iampb.User{
		FullName:     req.GetEmailOrUsername(),
		ControlPlane: typepb.ControlPlane_CONTROL_PLANE_ADMIN,
		Id:           "sentinez.admin",
	}, AccessToken: accessToken}, nil
}

func (srv *IAMService) Login(ctx context.Context,
	req *iampb.LoginRequest) (*iampb.LoginResponse, error) {

	if req.GetEmailOrUsername() ==
		srv.config.Get(settingpb.Senz_SENZ_ADMIN_USERNAME) {
		return srv.loginAdmin(req)
	}

	acc, err := srv.accounts.GetByProvider(
		ctx, iampb.Provider_PROVIDER_LOCAL, req.GetEmailOrUsername())
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

	accessToken, err := newAccessToken(user)
	if err != nil {
		return nil, err
	}

	return &iampb.LoginResponse{User: user, AccessToken: accessToken}, nil
}

func (srv *IAMService) CreateUser(ctx context.Context,
	request *iampb.CreateUserRequest) (*iampb.CreateUserResponse, error) {

	if err := srv.emailMustFree(ctx, request.GetEmail()); err != nil {
		return nil, err
	}

	user, err := srv.users.Create(ctx, &iampb.User{
		FullName:     request.GetFullName(),
		Email:        request.GetEmail(),
		ControlPlane: typepb.ControlPlane_CONTROL_PLANE_PORTAL,
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
		user, err := srv.users.GetByEmail(ctx, request.GetEmail())
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

	txss, err := srv.tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// accounts go with the user, or they keep its email taken
	err = srv.accounts.WithTX(txss).DeleteByUserId(ctx, request.GetId())
	if err != nil {
		_ = txss.Rollback(ctx)
		return nil, err
	}

	if err = srv.users.WithTX(txss).Delete(ctx, request.GetId()); err != nil {
		_ = txss.Rollback(ctx)
		return nil, err
	}

	if err = txss.Commit(ctx); err != nil {
		return nil, err
	}

	return &iampb.DeleteUserResponse{}, nil
}

func (srv *IAMService) UpdateUser(ctx context.Context,
	request *iampb.UpdateUserRequest) (*iampb.UpdateUserResponse, error) {

	acc, err := srv.findAccount(ctx, request.GetEmail())
	if err != nil {
		return nil, err
	}

	if acc == nil {
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
		dest.Email = req.GetEmail()
	}

	if req.GetFullName() != "" {
		dest.FullName = req.GetFullName()
	}
}
