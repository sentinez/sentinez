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

// Package iamdomains provides the
// Identity Access Management domain service implementation.
package iamdomains

import (
	"context"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/domain/v1"
	iamdomainpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/domain/v1"
)

var _ iamdomainpb.
	IdentityAccessManagerDomainServiceServer = (*IdentityAccessManager)(nil)

func New() iamdomainpb.IdentityAccessManagerDomainServiceServer {
	return &IdentityAccessManager{}
}

// IdentityAccessManager implements the Identity Access Manager domain service.
type IdentityAccessManager struct {
}

// CreateUser implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) CreateUser(
	ctx context.Context,
	req *iamdomainpb.CreateUserRequest,
) (*iamdomainpb.CreateUserResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// GetUser implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) GetUser(ctx context.Context,
	req *iamdomainpb.GetUserRequest) (*iamdomainpb.GetUserResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// ListUsers implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) ListUsers(ctx context.Context,
	req *iamdomainpb.ListUsersRequest) (*iamdomainpb.ListUsersResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// DeleteUser implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) DeleteUser(
	ctx context.Context,
	req *iamdomainpb.DeleteUserRequest,
) (*iamdomainpb.DeleteUserResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// UpdateUser implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) UpdateUser(
	ctx context.Context,
	req *iamdomainpb.UpdateUserRequest,
) (*iamdomainpb.UpdateUserResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// Status implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) Status(ctx context.Context,
	req *domain.StatusRequest) (*domain.StatusResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}
