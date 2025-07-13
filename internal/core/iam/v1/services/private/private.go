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

package privateservice

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"
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
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}

func (srv *IAMPrivateService) ListUsers(ctx context.Context,
	request *iam.ListUsersRequest) (*iam.ListUsersResponse, error) {
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}

func (srv *IAMPrivateService) DeleteUser(ctx context.Context,
	request *iam.DeleteUserRequest) (*iam.DeleteUserResponse, error) {
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}

func (srv *IAMPrivateService) UpdateUser(ctx context.Context,
	request *iam.UpdateUserRequest) (*iam.UpdateUserResponse, error) {
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}
