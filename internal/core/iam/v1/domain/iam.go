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

	usersrepo "github.com/sentinez/sentinez/internal/core/iam/v1/repos/users"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
)

var _ iam.
	IdentityAccessManagerDomainServiceServer = (*IdentityAccessManager)(nil)

func New(users usersrepo.IUser) iam.IdentityAccessManagerDomainServiceServer {
	return &IdentityAccessManager{
		users: users,
	}
}

// IdentityAccessManager implements the Identity Access Manager domain service.
type IdentityAccessManager struct {
	users usersrepo.IUser
}

// CreateUser implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) DomainCreateUser(
	ctx context.Context,
	req *iam.DomainCreateUserRequest,
) (*iam.DomainCreateUserResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// GetUser implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) DomainGetUser(ctx context.Context,
	req *iam.DomainGetUserRequest) (*iam.DomainGetUserResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// ListUsers implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) DomainListUsers(ctx context.Context,
	req *iam.DomainListUsersRequest) (*iam.DomainListUsersResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// DeleteUser implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) DomainDeleteUser(
	ctx context.Context,
	req *iam.DomainDeleteUserRequest,
) (*iam.DomainDeleteUserResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// UpdateUser implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) DomainUpdateUser(
	ctx context.Context,
	req *iam.DomainUpdateUserRequest,
) (*iam.DomainUpdateUserResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}

// Status implements domain.IdentityAccessManagerDomainServiceServer.
func (iam *IdentityAccessManager) DomainStatus(ctx context.Context,
	req *iam.DomainStatusRequest) (*iam.DomainStatusResponse, error) {

	_, _ = ctx, req

	//TODO implement me
	panic("implement me")
}
