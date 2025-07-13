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

package publicservice

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"
)

func New(users usersrepo.IUser) *IAMPublicService {
	return &IAMPublicService{
		users: users,
	}
}

type IAMPublicService struct {
	users usersrepo.IUser
}

func (srv *IAMPublicService) Status(ctx context.Context,
	request *iam.StatusRequest) (*iam.StatusResponse, error) {
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}

func (srv *IAMPublicService) CreateAccount(ctx context.Context,
	request *iam.CreateAccountRequest) (*iam.CreateAccountResponse, error) {
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}

func (srv *IAMPublicService) Login(ctx context.Context,
	request *iam.LoginRequest) (*iam.LoginResponse, error) {
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}
