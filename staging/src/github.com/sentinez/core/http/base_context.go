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

package corehttp

import "github.com/sentinez/shared/atomic"

var _ RequestContext = (*BaseContext)(nil)

type BaseContext struct {
	requestId  atomic.WriteOnce[string]
	resourceId atomic.WriteOnce[string]
	tenantId   atomic.WriteOnce[string]
}

// RequestId implements [RequestContext].
func (b *BaseContext) RequestId() string {
	id, _ := b.requestId.Get()
	return id
}

// ResourceId implements [RequestContext].
func (b *BaseContext) ResourceId() string {
	id, _ := b.resourceId.Get()
	return id
}

// SetRequestId implements [RequestContext].
func (b *BaseContext) SetRequestId(id string) {
	_ = b.requestId.Set(id)
}

// SetResourceId implements [RequestContext].
func (b *BaseContext) SetResourceId(id string) {
	_ = b.resourceId.Set(id)
}

// SetTenantId implements [RequestContext].
func (b *BaseContext) SetTenantId(id string) {
	_ = b.tenantId.Set(id)
}

// TenantId implements [RequestContext].
func (b *BaseContext) TenantId() string {
	id, _ := b.tenantId.Get()
	return id
}

func (b *BaseContext) Reset() {
	if b == nil {
		return
	}

	b.requestId.Reset()
	b.tenantId.Reset()
	b.resourceId.Reset()
}
