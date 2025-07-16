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

package iamprivservice

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"
	stderr "github.com/sentinez/sentinez/pkg/std/errors"
)

func New(users usersrepo.IUser) *IAMPrivateService {
	return &IAMPrivateService{
		users: users,
	}
}

type IAMPrivateService struct {
	users usersrepo.IUser
}

func (srv *IAMPrivateService) CreateUser(ctx context.Context,
	request *iam.CreateUserRequest) (*iam.CreateUserResponse, error) {

	user, err := srv.users.Create(ctx, &iam.Users{
		FullName:    request.FullName,
		Email:       request.Email,
		PhoneNumber: request.PhoneNumber,
	})
	if err != nil {
		return nil, err
	}

	return &iam.CreateUserResponse{UserId: user.Id}, nil
}

func (srv *IAMPrivateService) GetUser(ctx context.Context,
	request *iam.GetUserRequest) (*iam.GetUserResponse, error) {
	if request.GetId() != "" {
		user, err := srv.users.Get(ctx, request.GetId())
		if err != nil {
			return nil, err
		}
		return &iam.GetUserResponse{User: user}, nil
	}

	if request.GetUsername() != "" {
		user, err := srv.users.GetByUsernameOrEmail(ctx, request.GetUsername())
		if err != nil {
			return nil, err
		}

		return &iam.GetUserResponse{User: user}, nil
	}

	return nil, stderr.F("invalid argument: must provide either id or username")
}

func (srv *IAMPrivateService) ListUsers(ctx context.Context,
	request *iam.ListUsersRequest) (*iam.ListUsersResponse, error) {

	users, err := srv.users.List(ctx, request)
	if err != nil {
		return nil, err
	}
	return &iam.ListUsersResponse{Users: users}, nil
}

func (srv *IAMPrivateService) DeleteUser(ctx context.Context,
	request *iam.DeleteUserRequest) (*iam.DeleteUserResponse, error) {

	if err := srv.users.Delete(ctx, request.GetId()); err != nil {
		return nil, err
	}
	return &iam.DeleteUserResponse{}, nil
}

func (srv *IAMPrivateService) UpdateUser(ctx context.Context,
	request *iam.UpdateUserRequest) (*iam.UpdateUserResponse, error) {

	user, err := srv.users.GetByUsernameOrEmail(ctx, request.GetEmail())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if user.GetId() != "" || errors.Is(err, pgx.ErrNoRows) {
		return nil, stderr.F("email %s already exists", request.GetEmail())
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
