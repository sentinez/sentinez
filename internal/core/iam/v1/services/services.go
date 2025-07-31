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

package iamservices

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	accountrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"
	stderr "github.com/sentinez/sentinez/pkg/std/errors"
)

var _ iam.IdentityAccessManagementServiceServer = (*IAMService)(nil)

func New(
	users usersrepo.IUser,
	account accountrepo.IAccount,
) *IAMService {
	return &IAMService{
		users:    users,
		accounts: account,
	}
}

type IAMService struct {
	users    usersrepo.IUser
	accounts accountrepo.IAccount
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
	if stderr.NotRowsNotFound(err) {
		return err
	}
	if acc.GetId() != "" {
		return stderr.AlreadyExistsF(
			"username %s already exists", acc.GetUsername())
	}

	acc, err = srv.accounts.GetByUsernameOrEmail(ctx, email)
	if stderr.NotRowsNotFound(err) {
		return err
	}
	if acc.GetId() != "" {
		return stderr.AlreadyExistsF(
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

	user, err := srv.CreateUser(ctx, &iam.CreateUserRequest{
		FullName:    request.GetFullName(),
		Email:       request.GetEmail(),
		PhoneNumber: request.GetPhoneNumber(),
	})
	if err != nil {
		return nil, err
	}

	acc, err := srv.accounts.Create(ctx, &iam.Accounts{
		UserId:   user.GetUserId(),
		Email:    request.GetEmail(),
		Username: request.GetUsername(),
		Password: request.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return &iam.CreateAccountResponse{
		AccountId: acc.GetId(),
	}, nil
}

func (srv *IAMService) Login(ctx context.Context,
	request *iam.LoginRequest) (*iam.LoginResponse, error) {
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}

func (srv *IAMService) CreateUser(ctx context.Context,
	request *iam.CreateUserRequest) (*iam.CreateUserResponse, error) {

	user, err := srv.users.GetByUsernameOrEmail(ctx, request.GetEmail())
	if stderr.NotRowsNotFound(err) {
		return nil, err
	}

	if user.GetId() != "" {
		return nil, stderr.AlreadyExistsF(
			"email %s already exists", request.GetEmail())
	}

	user, err = srv.users.Create(ctx, &iam.Users{
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
		user, err := srv.users.GetByUsernameOrEmail(ctx, request.GetEmail())
		if err != nil {
			return nil, err
		}

		return &iam.GetUserResponse{User: user}, nil
	}

	return nil, stderr.InvalidDataF(
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

	user, err := srv.users.GetByUsernameOrEmail(ctx, request.GetEmail())
	if stderr.NotRowsNotFound(err) {
		return nil, err
	}

	if user.GetId() != "" {
		return nil, stderr.AlreadyExistsF(
			"email %s already exists", request.GetEmail())
	}

	_, err = srv.users.Update(ctx, &iam.Users{
		Id:          request.GetId(),
		FullName:    request.GetFullName(),
		Email:       request.GetEmail(),
		PhoneNumber: request.GetPhoneNumber(),
	})
	if err != nil {
		return nil, err
	}

	return &iam.UpdateUserResponse{}, nil
}
