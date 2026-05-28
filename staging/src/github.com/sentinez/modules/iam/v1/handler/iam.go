// Copyright 2025 Duc-Hung Ho.
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

// Package iamhdls provides the Identity Access Management service handler.
package iamhdl

import (
	"context"

	iamsvc "github.com/sentinez/modules/iam/v1/service"
	"github.com/sentinez/modules/pkg/headers"
	greeterpb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/greeter/v1"
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/iam/v1"
	"github.com/sentinez/shared/errorx"
	"github.com/sentinez/shared/zlog"
)

var _ iampb.
	IdentityAccessManagementServiceServer = (*IdentityAccessManagement)(nil)

func New(
	service *iamsvc.IAMService,
	greeterCli greeterpb.GreeterServiceClient,
) iampb.IdentityAccessManagementServiceServer {
	return &IdentityAccessManagement{
		service:    service,
		greeterCli: greeterCli,
	}
}

type IdentityAccessManagement struct {
	greeterCli greeterpb.GreeterServiceClient
	service    *iamsvc.IAMService
}

// PasskeyLoginVerify implements iam.IdentityAccessManagementServiceServer.
func (iam *IdentityAccessManagement) PasskeyLoginVerify(
	ctx context.Context,
	req *iampb.PasskeyLoginVerifyRequest,
) (*iampb.PasskeyLoginVerifyResponse, error) {

	return iam.service.PasskeyLoginVerify(ctx, req)
}

// PasskeyLoginChallenge implements iam.IdentityAccessManagementServiceServer.
func (iam *IdentityAccessManagement) PasskeyLoginChallenge(
	ctx context.Context,
	req *iampb.PasskeyLoginChallengeRequest,
) (*iampb.PasskeyLoginChallengeResponse, error) {

	return iam.service.PasskeyLoginChallenge(ctx, req)
}

// PasskeyRegisterVerify implements iam.IdentityAccessManagementServiceServer.
func (iam *IdentityAccessManagement) PasskeyRegisterVerify(
	ctx context.Context,
	req *iampb.PasskeyRegisterVerifyRequest,
) (*iampb.PasskeyRegisterVerifyResponse, error) {
	zlog.Debug("[IdentityAccessManagement.PasskeyRegisterVerify]")

	return iam.service.PasskeyRegisterVerify(ctx, req)
}

// PasskeyRegisterChallenge implements iam.IdentityAccessManagementServiceServer
func (iam *IdentityAccessManagement) PasskeyRegisterChallenge(
	ctx context.Context,
	req *iampb.PasskeyRegisterChallengeRequest,
) (*iampb.PasskeyRegisterChallengeResponse, error) {
	zlog.Debugf("[IAMMNT.PasskeyRegisterChallenge] req = %v", req)

	user, err := iam.service.
		GetAccountByUsernameOrEmail(ctx, req.GetEmailOrUsername())
	if err != nil {
		return nil, err
	}

	if user.GetId() != "" {
		return nil, errorx.StatusAlreadyExistsF(
			"username or email already exists: %s", req.GetEmailOrUsername())
	}

	return iam.service.PasskeyRegisterChallenge(ctx, req)
}

func (iam *IdentityAccessManagement) ListAccounts(ctx context.Context,
	request *iampb.ListAccountsRequest) (*iampb.ListAccountsResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.ListAccounts] req = %v", request)

	ss, err := headers.GetAuth(ctx)
	if err != nil {
		return nil, err
	}

	err = ss.Check(iampb.GetIdentityAccessManagementServiceListUsers())
	if err != nil {
		return nil, err
	}

	resp, err := iam.service.ListAccounts(ctx, request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (iam *IdentityAccessManagement) GetUser(ctx context.Context,
	request *iampb.GetUserRequest) (*iampb.GetUserResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.GetUser] request= %v", request)

	resp, err := iam.service.GetUser(ctx, request)
	if err != nil {
		zlog.Errorf("failed to get users: %v", err)
		return nil, err
	}

	return resp, nil
}

func (iam *IdentityAccessManagement) ListUsers(ctx context.Context,
	request *iampb.ListUsersRequest) (*iampb.ListUsersResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.ListUsers] request= %v", request)

	resp, err := iam.service.ListUsers(ctx, request)
	if err != nil {
		zlog.Errorf("failed to list users: %v", err)
		return nil, err
	}

	return resp, nil
}

func (iam *IdentityAccessManagement) DeleteUser(ctx context.Context,
	request *iampb.DeleteUserRequest) (*iampb.DeleteUserResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.DeleteUser] request= %v", request)

	resp, err := iam.service.DeleteUser(ctx, request)
	if err != nil {
		zlog.Errorf("faild to delete users: %s", err)
		return nil, err
	}

	return resp, nil
}

func (iam *IdentityAccessManagement) UpdateUser(ctx context.Context,
	request *iampb.UpdateUserRequest) (*iampb.UpdateUserResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.UpdateUser] request= %v", request)

	resp, err := iam.service.UpdateUser(ctx, request)
	if err != nil {
		zlog.Errorf("[IdentityAccessManagement.UpdateUser] update err: %v", err)
		return nil, err
	}

	return resp, nil
}

func (iam *IdentityAccessManagement) CreateAccount(ctx context.Context,
	req *iampb.CreateAccountRequest) (*iampb.CreateAccountResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.CreateAccount] request= %v", req)

	accResp, err := iam.service.CreateAccount(ctx, &iampb.CreateAccountRequest{
		Username:    req.GetUsername(),
		Password:    req.GetPassword(),
		Email:       req.GetEmail(),
		PhoneNumber: req.GetPhoneNumber(),
		FullName:    req.GetFullName(),
	})
	if err != nil {
		zlog.Errorf("failed to create account: %v", err)
		return nil, err
	}

	return &iampb.CreateAccountResponse{
		AccountId: accResp.GetAccountId(),
	}, nil
}

func (iam *IdentityAccessManagement) CreateUser(ctx context.Context,
	req *iampb.CreateUserRequest) (*iampb.CreateUserResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.CreateUser] request= %v", req)

	resp, err := iam.service.CreateUser(ctx, req)
	if err != nil {
		zlog.Errorf("failed to create user: %v", err)
		return nil, err
	}

	return resp, nil
}

func (iam *IdentityAccessManagement) Login(ctx context.Context,
	req *iampb.LoginRequest) (*iampb.LoginResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.Login] username = %s",
		req.GetEmailOrUsername())

	resp, err := iam.service.Login(ctx, req)
	if err != nil {
		zlog.Errorf("IAM.Login: failed to login user err=%v", err)
		return nil, err
	}

	return resp, nil
}

func (iam *IdentityAccessManagement) Status(ctx context.Context,
	req *iampb.StatusRequest) (*iampb.StatusResponse, error) {
	zlog.Debugf("request= %v", req)

	_, err := iam.greeterCli.Status(ctx, &greeterpb.StatusRequest{})
	if err != nil {
		zlog.Errorf("IAM.Status call greeter err=%v", err)
		return nil, err
	}

	ss, _ := headers.GetAuth(ctx)

	return &iampb.StatusResponse{
		Msg:     "OK",
		Context: ss.Context,
	}, nil
}
