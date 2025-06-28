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

// Package iamhandlers provides the Identity Access Management service handler.
package iamhandler

import (
	"context"

	iampb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var _ iampb.
	IdentityAccessManagementServiceServer = (*IdentityAccessManagement)(nil)

func New(domain iampb.IdentityAccessManagerDomainServiceServer,
) iampb.IdentityAccessManagementServiceServer {
	return &IdentityAccessManagement{
		domain: domain,
	}
}

// IdentityAccessManagement implement iampb.IAMServiceServer
type IdentityAccessManagement struct {
	domain iampb.IdentityAccessManagerDomainServiceServer
}

// CreateAccount implements iampb.IAMServiceServer.
func (iam *IdentityAccessManagement) CreateAccount(ctx context.Context,
	req *iampb.CreateAccountRequest) (*iampb.CreateAccountResponse, error) {

	_, _ = ctx, req

	panic("unimplemented")
}

// CreateUser implements iampb.IAMServiceServer.
func (iam *IdentityAccessManagement) CreateUser(ctx context.Context,
	req *iampb.CreateUserRequest) (*iampb.CreateUserResponse, error) {

	_, _ = ctx, req

	panic("unimplemented")
}

// Login implements iampb.IAMServiceServer.
func (iam *IdentityAccessManagement) Login(ctx context.Context,
	req *iampb.LoginRequest) (*iampb.LoginResponse, error) {

	_, _ = ctx, req

	panic("unimplemented")
}

// Status implements iampb.IAMServiceServer.
func (iam *IdentityAccessManagement) Status(ctx context.Context,
	req *iampb.StatusRequest) (*iampb.StatusResponse, error) {
	zlog.Debugf("request= %v", req)

	_ = ctx

	return &iampb.StatusResponse{
		Msg: "OK",
	}, nil
}
