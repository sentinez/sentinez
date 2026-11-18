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

// Package auditowner carries the owner of an audited call, the tenant and
// the resource it is about, from the checks of its access to its audit
// log: the checks read them from the database, where the request may only
// name an id of its own.
package auditowner

import (
	"context"
	"sync"
)

type ownerKey struct{}

// Owner the tenant and the resource of a call, the first ones the checks
// of its access looked at.
type Owner struct {
	mu         sync.Mutex
	tenantId   string
	resourceId string
}

// With returns ctx carrying a new Owner, which Set fills, and it.
func With(ctx context.Context) (context.Context, *Owner) {
	owner := &Owner{}
	return context.WithValue(ctx, ownerKey{}, owner), owner
}

// Set sets the unset tenant and resource of the Owner of ctx to tenantId
// and resourceId, "" leaving one unset; nothing without an Owner.
func Set(ctx context.Context, tenantId, resourceId string) {
	if owner, ok := ctx.Value(ownerKey{}).(*Owner); ok {
		owner.Set(tenantId, resourceId)
	}
}

// Set sets the unset tenant and resource of o to tenantId and resourceId,
// "" leaving one unset.
func (o *Owner) Set(tenantId, resourceId string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.tenantId == "" {
		o.tenantId = tenantId
	}
	if o.resourceId == "" {
		o.resourceId = resourceId
	}
}

// Ids returns the tenant and the resource of o, "" for one never set.
func (o *Owner) Ids() (tenantId, resourceId string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.tenantId, o.resourceId
}
