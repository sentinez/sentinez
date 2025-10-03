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

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	iamsvc "github.com/sentinez/sentinez/internal/core/iam/v1/services"
	"github.com/sentinez/sentinez/pkg/std/stdctx"
	"github.com/sentinez/sentinez/pkg/std/stderr"
	"github.com/sentinez/sentinez/pkg/std/stdperms"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var _ iampb.
	IdentityAccessManagementServiceServer = (*IdentityAccessManagement)(nil)

func New(
	service *iamsvc.IAMService,
	greeterCli greeter.GreeterServiceClient,
) iampb.IdentityAccessManagementServiceServer {
	return &IdentityAccessManagement{
		service:    service,
		greeterCli: greeterCli,
	}
}

type IdentityAccessManagement struct {
	greeterCli greeter.GreeterServiceClient
	service    *iamsvc.IAMService
}

// PasskeyLoginFinish implements iam.IdentityAccessManagementServiceServer.
func (iam *IdentityAccessManagement) PasskeyLoginFinish(
	ctx context.Context,
	req *iampb.PasskeyLoginFinishRequest,
) (*iampb.PasskeyLoginFinishResponse, error) {
	panic("unimplemented")
}

// PasskeyLoginStart implements iam.IdentityAccessManagementServiceServer.
func (iam *IdentityAccessManagement) PasskeyLoginStart(
	ctx context.Context,
	req *iampb.PasskeyLoginStartRequest,
) (*iampb.PasskeyLoginStartResponse, error) {
	panic("unimplemented")
}

// PasskeyRegisterFinish implements iam.IdentityAccessManagementServiceServer.
func (iam *IdentityAccessManagement) PasskeyRegisterFinish(
	ctx context.Context,
	req *iampb.PasskeyRegisterFinishRequest,
) (*iampb.PasskeyRegisterFinishResponse, error) {
	zlog.Debug("[IdentityAccessManagement.PasskeyRegisterFinish]")

	return iam.service.PasskeyRegisterFinish(ctx, req)
}

// PasskeyRegisterStart implements iam.IdentityAccessManagementServiceServer.
func (iam *IdentityAccessManagement) PasskeyRegisterStart(
	ctx context.Context,
	req *iampb.PasskeyRegisterStartRequest,
) (*iampb.PasskeyRegisterStartResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.PasskeyRegisterStart] req = %v", req)

	user, err := iam.service.
		GetAccountByUsernameOrEmail(ctx, req.GetEmailOrUsername())
	if err != nil {
		return nil, err
	}

	if user.GetId() != "" {
		return nil, stderr.AlreadyExistsF(
			"username or email already exists: %s", req.GetEmailOrUsername())
	}

	return iam.service.PasskeyRegisterStart(ctx, req)
}

func (iam *IdentityAccessManagement) ListAccounts(ctx context.Context,
	request *iampb.ListAccountsRequest) (*iampb.ListAccountsResponse, error) {
	zlog.Debugf("[IdentityAccessManagement.ListAccounts] req = %v", request)

	ss, err := stdctx.GetAuthContext(ctx, iam.service.Config())
	if err != nil {
		return nil, err
	}

	if !stdperms.HasLeastOne(
		ss.GetPermissionBitwise(), stdperms.DefaultViewAny()) {
		request.UserIds = []string{ss.GetUserId()}
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
		zlog.Errorf("[IdentityAccessManagement.UpdateUser] update err=", err)
		return nil, err
	}

	return resp, nil
}

// CreateAccount implements iampb.iamsvcerver.
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

// CreateUser implements iampb.iamsvcerver.
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

// Login implements iampb.iamsvcerver.
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

// Status implements iampb.iamsvcerver.
func (iam *IdentityAccessManagement) Status(ctx context.Context,
	req *iampb.StatusRequest) (*iampb.StatusResponse, error) {
	zlog.Debugf("request= %v", req)

	_, err := iam.greeterCli.Status(ctx, &greeter.StatusRequest{})
	if err != nil {
		zlog.Errorf("IAM.Status call greeter err=%v", err)
		return nil, err
	}

	ss, _ := stdctx.GetAuthContext(ctx, iam.service.Config())

	return &iampb.StatusResponse{
		Msg:     "OK",
		Context: ss,
	}, nil
}
