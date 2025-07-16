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

package iampubservice

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	accountrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/accounts"
	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"
)

func New(
	users usersrepo.IUser,
	account accountrepo.IAccount,
) *IAMPublicService {
	return &IAMPublicService{
		users:    users,
		accounts: account,
	}
}

type IAMPublicService struct {
	users    usersrepo.IUser
	accounts accountrepo.IAccount
}

func (srv *IAMPublicService) Status(ctx context.Context,
	request *iam.StatusRequest) (*iam.StatusResponse, error) {
	_ = ctx
	_ = request

	return &iam.StatusResponse{Msg: "OK"}, nil
}

func (srv *IAMPublicService) CreateAccount(ctx context.Context, userID string,
	request *iam.CreateAccountRequest) (*iam.CreateAccountResponse, error) {

	acc, err := srv.accounts.Create(ctx, &iam.Accounts{
		UserId:   userID,
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

func (srv *IAMPublicService) Login(ctx context.Context,
	request *iam.LoginRequest) (*iam.LoginResponse, error) {
	_ = ctx
	_ = request
	//TODO implement me
	panic("implement me")
}
