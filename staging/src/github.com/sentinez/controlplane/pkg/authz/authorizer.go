// Copyright 2026 Duc-Hung Ho.
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

package authz

import (
	"context"
	"fmt"
	"slices"

	"github.com/sentinez/controlplane/pkg/headers"
	"github.com/sentinez/pkg/szerror"
	tenantpb "github.com/sentinez/sentinez/api/proto/sentinez/apps/tenant/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The least role of a tenant for each kind of call.
const (
	// Read lets the caller see a tenant and what belongs to it.
	Read = typepb.Role_ROLE_MEMBER
	// Write lets the caller change the resources and the security rules.
	Write = typepb.Role_ROLE_LEADER
	// Manage lets the caller change the tenant and its members.
	Manage = typepb.Role_ROLE_ADMIN
)

// Rank orders the roles: the higher, the more the role can do. An unknown
// role ranks 0.
func Rank(role typepb.Role) int {
	switch role {
	case typepb.Role_ROLE_SUPER_ADMIN:
		return 4
	case typepb.Role_ROLE_ADMIN:
		return 3
	case typepb.Role_ROLE_LEADER:
		return 2
	case typepb.Role_ROLE_MEMBER:
		return 1
	default:
		return 0
	}
}

// RankOf returns the highest rank of roles.
func RankOf(roles []typepb.Role) int {
	best := 0
	for _, role := range roles {
		best = max(best, Rank(role))
	}

	return best
}

// MemberLister lists the members of the tenants.
type MemberLister interface {
	List(ctx context.Context, req *tenantpb.ListMemberRequest,
	) (*tenantpb.ListMemberResponse, error)
}

// ResourceReader reads the resources of the tenants.
type ResourceReader interface {
	Get(ctx context.Context, id string) (*tenantpb.Resource, error)
	List(ctx context.Context, req *tenantpb.ListResourceRequest,
	) (*tenantpb.ListResourceResponse, error)
}

// Scope is what a list call may look at.
type Scope struct {
	// Ids filter the call; empty means it needs no filter.
	Ids []string
	// None is true when the caller may look at nothing: the call must not
	// run, as an empty filter would match everything.
	None bool
}

// Authorizer decides what a signed-in caller may do in the tenants.
//
// The admin control plane may do anything. Any other caller acts through
// the roles of its memberships, read from the database on every call, so a
// change of role applies at once.
type Authorizer struct {
	members   MemberLister
	resources ResourceReader
}

// New returns an Authorizer reading the memberships from members and the
// resources from resources.
func New(members MemberLister, resources ResourceReader) *Authorizer {
	return &Authorizer{members: members, resources: resources}
}

// RequireTenant fails unless the caller has at least role least in the
// tenant. A caller outside the tenant gets NotFound, so that it learns
// nothing of the tenant; one inside it without the role gets
// PermissionDenied.
func (a *Authorizer) RequireTenant(ctx context.Context, tenantId string,
	least typepb.Role) error {

	_, err := a.rankIn(ctx, tenantId, least)

	return err
}

// rankIn is RequireTenant that also returns the rank of the caller in the
// tenant; the admin control plane ranks above every role.
func (a *Authorizer) rankIn(ctx context.Context, tenantId string,
	least typepb.Role) (int, error) {

	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return 0, err
	}

	if auth.IsAdmin() {
		return Rank(typepb.Role_ROLE_SUPER_ADMIN), nil
	}

	userId, err := userIdOf(auth)
	if err != nil {
		return 0, err
	}

	rank, err := a.memberRank(ctx, tenantId, userId)
	if err != nil {
		return 0, err
	}

	if rank < Rank(least) {
		return 0, szerror.StatusForbiddenF(
			"role %s is required in tenant %s", least, tenantId)
	}

	return rank, nil
}

// memberRank returns the highest rank of the user in the tenant, NotFound
// for a user who is not a member.
func (a *Authorizer) memberRank(ctx context.Context, tenantId,
	userId string) (int, error) {

	resp, err := a.members.List(ctx, &tenantpb.ListMemberRequest{
		TenantIds: []string{tenantId},
		UserId:    userId,
	})
	if err != nil {
		return 0, fmt.Errorf("member.List: %w", err)
	}

	if len(resp.GetMembers()) == 0 {
		return 0, szerror.StatusNotFoundF("tenant not found: id=%s", tenantId)
	}

	rank := 0
	for _, member := range resp.GetMembers() {
		rank = max(rank, RankOf(member.GetRoles()))
	}

	return rank, nil
}

// RequireResource fails unless the caller has at least role least in the
// tenant of the resource, and returns the resource. A caller outside the
// tenant gets NotFound.
func (a *Authorizer) RequireResource(ctx context.Context, resourceId string,
	least typepb.Role) (*tenantpb.Resource, error) {

	resource, err := a.resources.Get(ctx, resourceId)
	if szerror.IsNoRows(err) {
		return nil, notFoundResource(resourceId)
	}
	if err != nil {
		return nil, fmt.Errorf("resource.Get: %w", err)
	}

	if err = a.RequireResourceOf(ctx, resource, least); err != nil {
		return nil, err
	}

	return resource, nil
}

// RequireResourceOf is RequireResource for a resource already read. The
// error of a caller outside the tenant does not name the tenant.
func (a *Authorizer) RequireResourceOf(ctx context.Context,
	resource *tenantpb.Resource, least typepb.Role) error {

	err := a.RequireTenant(ctx, resource.GetTenantId(), least)
	if status.Code(err) == codes.NotFound {
		return notFoundResource(resource.GetId())
	}

	return err
}

// RequireGrant fails unless the caller may give roles to a member of the
// tenant: it must manage the tenant, and may not give a role above its own
// nor the super admin role, which only the admin control plane gives.
func (a *Authorizer) RequireGrant(ctx context.Context, tenantId string,
	roles []typepb.Role) error {

	rank, err := a.rankIn(ctx, tenantId, Manage)
	if err != nil {
		return err
	}

	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if role == typepb.Role_ROLE_SUPER_ADMIN && !auth.IsAdmin() ||
			Rank(role) > rank {

			return szerror.StatusForbiddenF("role %s can not be given", role)
		}
	}

	return nil
}

// TenantIds limits requested to the tenants the caller is a member of;
// without requested, to all of them.
func (a *Authorizer) TenantIds(ctx context.Context, requested []string,
) (Scope, error) {

	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return Scope{}, err
	}

	if auth.IsAdmin() {
		return Scope{Ids: requested}, nil
	}

	own, err := a.tenantIdsOf(ctx, auth)
	if err != nil {
		return Scope{}, err
	}

	return narrow(requested, own), nil
}

// ResourceIds limits requested to the resources of the tenants the caller
// is a member of; without requested, to all of them.
func (a *Authorizer) ResourceIds(ctx context.Context, requested []string,
) (Scope, error) {

	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return Scope{}, err
	}

	if auth.IsAdmin() {
		return Scope{Ids: requested}, nil
	}

	tenantIds, err := a.tenantIdsOf(ctx, auth)
	if err != nil || len(tenantIds) == 0 {
		return Scope{None: true}, err
	}

	resp, err := a.resources.List(ctx,
		&tenantpb.ListResourceRequest{TenantIds: tenantIds})
	if err != nil {
		return Scope{}, fmt.Errorf("resource.List: %w", err)
	}

	own := make([]string, 0, len(resp.GetResources()))
	for _, resource := range resp.GetResources() {
		own = append(own, resource.GetId())
	}

	return narrow(requested, own), nil
}

// CoMemberUserIds returns the ids of the users sharing a tenant with the
// caller, itself included.
func (a *Authorizer) CoMemberUserIds(ctx context.Context) ([]string, error) {
	auth, err := headers.GetAuth(ctx)
	if err != nil {
		return nil, err
	}

	userId, err := userIdOf(auth)
	if err != nil {
		return nil, err
	}

	userIds := []string{userId}

	tenantIds, err := a.tenantIdsOf(ctx, auth)
	if err != nil || len(tenantIds) == 0 {
		return userIds, err
	}

	resp, err := a.members.List(ctx,
		&tenantpb.ListMemberRequest{TenantIds: tenantIds})
	if err != nil {
		return nil, fmt.Errorf("member.List: %w", err)
	}

	for _, member := range resp.GetMembers() {
		if !slices.Contains(userIds, member.GetUserId()) {
			userIds = append(userIds, member.GetUserId())
		}
	}

	return userIds, nil
}

func (a *Authorizer) tenantIdsOf(ctx context.Context, auth *headers.Auth,
) ([]string, error) {
	userId, err := userIdOf(auth)
	if err != nil {
		return nil, err
	}

	resp, err := a.members.List(ctx,
		&tenantpb.ListMemberRequest{UserId: userId})
	if err != nil {
		return nil, fmt.Errorf("member.List: %w", err)
	}

	tenantIds := make([]string, 0, len(resp.GetMembers()))
	for _, member := range resp.GetMembers() {
		tenantIds = append(tenantIds, member.GetTenantId())
	}

	return tenantIds, nil
}

// userIdOf returns the user of the token. Without it a member filter on
// the user would be dropped and match every member.
func userIdOf(auth *headers.Auth) (string, error) {
	if auth.GetUserId() == "" {
		return "", szerror.StatusUnauthorizedF("token has no user")
	}

	return auth.GetUserId(), nil
}

// narrow keeps the requested ids that are in allowed; without requested,
// all of allowed.
func narrow(requested, allowed []string) Scope {
	ids := allowed
	if len(requested) != 0 {
		ids = make([]string, 0, len(requested))
		for _, id := range requested {
			if slices.Contains(allowed, id) {
				ids = append(ids, id)
			}
		}
	}

	if len(ids) == 0 {
		return Scope{None: true}
	}

	return Scope{Ids: ids}
}

func notFoundResource(id string) error {
	return szerror.StatusNotFoundF("resource not found: id=%s", id)
}
