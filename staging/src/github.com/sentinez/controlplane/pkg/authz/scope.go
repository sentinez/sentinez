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

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// The fields of the requests that list by resource or by tenant.
const (
	_resourceIds protoreflect.Name = "resource_ids"
	_tenantIds   protoreflect.Name = "tenant_ids"
)

// ResourceScoped is a request that lists by resource: its resource_ids
// field tells which ones, none meaning all.
type ResourceScoped interface {
	proto.Message
	GetResourceIds() []string
}

// TenantScoped is a request that lists by tenant: its tenant_ids field
// tells which ones, none meaning all.
type TenantScoped interface {
	proto.Message
	GetTenantIds() []string
}

// Scoped returns a copy of req limited to the resources the caller may see.
// It returns false when the caller may see none: the call must then answer
// empty without running, as a request without resource ids matches all.
func Scoped[R ResourceScoped](ctx context.Context, a *Authorizer, req R,
) (R, bool, error) {
	scope, err := a.ResourceIds(ctx, req.GetResourceIds())

	return withIds(req, _resourceIds, scope, err)
}

// ScopedTenants is Scoped for the tenants of the caller.
func ScopedTenants[R TenantScoped](ctx context.Context, a *Authorizer, req R,
) (R, bool, error) {
	scope, err := a.TenantIds(ctx, req.GetTenantIds())

	return withIds(req, _tenantIds, scope, err)
}

func withIds[R proto.Message](req R, field protoreflect.Name, scope Scope,
	err error) (R, bool, error) {

	if err != nil || scope.None {
		return req, false, err
	}

	scoped := proto.CloneOf(req)
	m := scoped.ProtoReflect()
	list := m.Mutable(m.Descriptor().Fields().ByName(field)).List()

	list.Truncate(0)
	for _, id := range scope.Ids {
		list.Append(protoreflect.ValueOfString(id))
	}

	return scoped, true, nil
}
