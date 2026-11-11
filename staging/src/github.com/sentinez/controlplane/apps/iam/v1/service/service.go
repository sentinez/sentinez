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
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	accrepos "github.com/sentinez/controlplane/apps/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/controlplane/apps/iam/v1/repos/users"
	memberrepo "github.com/sentinez/controlplane/apps/tenant/v1/repos/members"
	"github.com/sentinez/controlplane/pkg/authz"
	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/sentinez/controlplane/pkg/headers"
	"github.com/sentinez/controlplane/pkg/passkey"
	"github.com/sentinez/pkg/szerror"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szstorage/dbx/postgres"
	iampb "github.com/sentinez/sentinez/api/proto/sentinez/apps/iam/v1"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ iampb.IdentityAccessManagementServiceServer = (*IAMService)(nil)

func New(config *settingpb.Config,
	tx *postgres.Tx,
	store passkey.Store,
	users usersrepo.IUser,
	account accrepos.IAccount,
	members memberrepo.IMember,
	az *authz.Authorizer,
) *IAMService {

	wauth := passkey.NewWebAuthn(config)

	return &IAMService{
		config:   config,
		tx:       tx,
		users:    users,
		accounts: account,
		members:  members,
		authz:    az,
		limits:   newSignInLimits(),
		store:    store,
		auth:     wauth,
	}
}

type IAMService struct {
	config   *settingpb.Config
	tx       *postgres.Tx
	users    usersrepo.IUser
	accounts accrepos.IAccount
	members  memberrepo.IMember
	authz    *authz.Authorizer
	limits   *signInLimits
	store    passkey.Store
	auth     *webauthn.WebAuthn
}

func (srv *IAMService) ListAccounts(ctx context.Context,
	request *iampb.ListAccountsRequest) (*iampb.ListAccountsResponse, error) {

	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return nil, err
	}

	// the accounts hold the sign-in methods: each user sees its own
	if !auth.IsAdmin() {
		scoped, ok := ownAccounts(request, auth.GetUserId())
		if !ok {
			return &iampb.ListAccountsResponse{}, nil
		}

		request = scoped
	}

	resp, err := srv.accounts.List(ctx, request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ownAccounts limits req to the accounts of userId. It returns false when
// req asks for the accounts of others only.
func ownAccounts(req *iampb.ListAccountsRequest, userId string,
) (*iampb.ListAccountsRequest, bool) {
	if ids := req.GetUserIds(); len(ids) != 0 && !slices.Contains(ids, userId) {
		return nil, false
	}

	scoped := proto.CloneOf(req)
	scoped.UserIds = []string{userId}

	return scoped, true
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
		return szerror.StatusAlreadyExistsF("email %s already exists", email)
	}

	return nil
}

func (srv *IAMService) usernameMustFree(
	ctx context.Context, username string) error {
	// the admin signs in with this name and has no account row
	if username == srv.config.Get(settingpb.Senz_SENZ_ADMIN_USERNAME) {
		return szerror.StatusAlreadyExistsF(
			"username %s already exists", username)
	}

	acc, err := srv.findLocalAccount(ctx, username)
	if err != nil {
		return err
	}

	if acc != nil {
		return szerror.StatusAlreadyExistsF(
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

// found turns szerror.ErrNotFound into a nil result.
func found[T any](v *T, err error) (*T, error) {
	if errors.Is(err, szerror.ErrNotFound) {
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
		return szerror.StatusAlreadyExistsF("email %s already exists", email)
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

	if !srv.isAdminPassword(req.GetPassword()) {
		return nil, szerror.StatusUnauthorizedF(
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

// login checks the credentials; Login limits how often it may fail.
func (srv *IAMService) login(ctx context.Context,
	req *iampb.LoginRequest) (*iampb.LoginResponse, error) {

	if req.GetEmailOrUsername() ==
		srv.config.Get(settingpb.Senz_SENZ_ADMIN_USERNAME) {
		return srv.loginAdmin(req)
	}

	acc, err := srv.accounts.GetByProvider(
		ctx, iampb.Provider_PROVIDER_LOCAL, req.GetEmailOrUsername())
	if err != nil {
		szlog.Debugf("faild to get account by username or email")
		return nil, err
	}

	if !crypto.CheckPasswordHash(req.GetPassword(), acc.GetPasswordHash()) {
		return nil,
			szerror.StatusUnauthorizedF("username, email or password is wrong!")
	}

	user, err := srv.users.Get(ctx, acc.GetUserId())
	if err != nil {
		szlog.Debugf("faild to get user by username or email")
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
		if err := srv.requireVisible(ctx, request.GetId()); err != nil {
			return nil, err
		}

		user, err := srv.users.Get(ctx, request.GetId())
		if err != nil {
			return nil, err
		}
		return &iampb.GetUserResponse{User: user}, nil
	}

	return nil, szerror.StatusInvalidDataF(
		"invalid argument: must provide either id or username")
}

func (srv *IAMService) ListUsers(ctx context.Context,
	request *iampb.ListUsersRequest) (*iampb.ListUsersResponse, error) {

	request, ok, err := srv.visibleUsers(ctx, request)
	if err != nil || !ok {
		return &iampb.ListUsersResponse{}, err
	}

	users, err := srv.users.List(ctx, request)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (srv *IAMService) DeleteUser(ctx context.Context,
	request *iampb.DeleteUserRequest) (*iampb.DeleteUserResponse, error) {

	// an empty id would match every member below
	if request.GetId() == "" {
		return nil, szerror.StatusInvalidArgumentF("id is required")
	}

	txss, err := srv.tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// a user who is gone is no member of a tenant any more
	if err = srv.deleteMemberships(ctx, txss, request.GetId()); err != nil {
		_ = txss.Rollback(ctx)
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

// UpdateUser renames the signed-in user; nothing else of the user changes.
func (srv *IAMService) UpdateUser(ctx context.Context,
	request *iampb.UpdateUserRequest) (*iampb.UpdateUserResponse, error) {

	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return nil, err
	}

	fullName := strings.TrimSpace(request.GetFullName())
	if fullName == "" {
		return nil, szerror.StatusInvalidArgumentF("full name is required")
	}

	user, err := srv.users.Get(ctx, auth.GetUserId())
	if err != nil {
		return nil, err
	}

	// the repo writes only the fields set here
	err = srv.users.Update(ctx,
		&iampb.User{Id: user.GetId(), FullName: fullName})
	if err != nil {
		return nil, err
	}

	user.FullName = fullName

	return &iampb.UpdateUserResponse{User: user}, nil
}

// deleteMemberships removes the user from every tenant.
func (srv *IAMService) deleteMemberships(ctx context.Context,
	txss *postgres.TxSession, userId string) error {

	members := srv.members.WithTX(txss)

	resp, err := members.List(ctx,
		&tenantpb.ListMemberRequest{UserId: userId})
	if err != nil {
		return err
	}

	for _, member := range resp.GetMembers() {
		if err = members.Delete(ctx, member.GetId()); err != nil {
			return err
		}
	}

	return nil
}

// _maxEmails is how many emails a user may look up at once.
const _maxEmails = 5

// requireVisible fails with NotFound unless the caller may see the user:
// itself, a user it shares a tenant with, or anyone for the admin control
// plane.
func (srv *IAMService) requireVisible(ctx context.Context, userId string,
) error {
	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return err
	}

	if auth.IsAdmin() || auth.GetUserId() == userId {
		return nil
	}

	coMembers, err := srv.authz.CoMemberUserIds(ctx)
	if err != nil {
		return err
	}

	if !slices.Contains(coMembers, userId) {
		return szerror.StatusNotFoundF("user not found: id=%s", userId)
	}

	return nil
}

// visibleUsers limits req to the users the caller may see; it returns
// false when there are none. The admin control plane sees all of them. Any
// other caller sees itself and the users of its tenants, and finds a user
// outside them only by email, to invite it: never by listing everyone.
func (srv *IAMService) visibleUsers(ctx context.Context,
	req *iampb.ListUsersRequest) (*iampb.ListUsersRequest, bool, error) {

	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return nil, false, err
	}

	if auth.IsAdmin() {
		return req, true, nil
	}

	if len(req.GetEmails()) > _maxEmails {
		return nil, false, szerror.StatusInvalidArgumentF(
			"at most %d emails can be looked up at once", _maxEmails)
	}

	scoped := proto.CloneOf(req)
	if len(req.GetIds()) == 0 {
		// a lookup by email goes through, no filter lists only itself
		if len(req.GetEmails()) == 0 {
			scoped.Ids = []string{auth.GetUserId()}
		}

		return scoped, true, nil
	}

	coMembers, err := srv.authz.CoMemberUserIds(ctx)
	if err != nil {
		return nil, false, err
	}

	visible := make(map[string]struct{}, len(coMembers))
	for _, id := range coMembers {
		visible[id] = struct{}{}
	}

	scoped.Ids = slices.DeleteFunc(slices.Clone(req.GetIds()),
		func(id string) bool { _, ok := visible[id]; return !ok })

	return scoped, len(scoped.GetIds()) != 0, nil
}

// isAdminPassword compares in constant time, and refuses when no admin
// password is configured.
func (srv *IAMService) isAdminPassword(password string) bool {
	want := srv.config.Get(settingpb.Senz_SENZ_ADMIN_PASSWORD)
	if want == "" {
		return false
	}

	// hashing gives both sides the same length
	got := sha256.Sum256([]byte(password))
	wantSum := sha256.Sum256([]byte(want))

	return subtle.ConstantTimeCompare(got[:], wantSum[:]) == 1
}
