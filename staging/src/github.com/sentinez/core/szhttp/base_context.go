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

package szhttp

import "github.com/sentinez/pkg/szatomic"

var _ RequestContext = (*BaseContext)(nil)

type BaseContext struct {
	requestId  szatomic.WriteOnce[string]
	resourceId szatomic.WriteOnce[string]
	tenantId   szatomic.WriteOnce[string]

	// The ids until set, kept by Reset.
	defaultResourceId string
	defaultTenantId   string
}

// NewBaseContext returns a BaseContext whose TenantId and ResourceId are
// tenantId and resourceId until they are set.
func NewBaseContext(tenantId, resourceId string) *BaseContext {
	return &BaseContext{
		defaultResourceId: resourceId,
		defaultTenantId:   tenantId,
	}
}

// RequestId implements [RequestContext].
func (b *BaseContext) RequestId() string {
	id, _ := b.requestId.Get()
	return id
}

// ResourceId implements [RequestContext].
func (b *BaseContext) ResourceId() string {
	if id, ok := b.resourceId.Get(); ok {
		return id
	}

	return b.defaultResourceId
}

// SetRequestId implements [RequestContext].
func (b *BaseContext) SetRequestId(id string) {
	_ = b.requestId.Set(id)
}

// SetResourceId implements [RequestContext]; an empty id keeps the
// default one.
func (b *BaseContext) SetResourceId(id string) {
	if id != "" {
		_ = b.resourceId.Set(id)
	}
}

// SetTenantId implements [RequestContext]; an empty id keeps the default
// one.
func (b *BaseContext) SetTenantId(id string) {
	if id != "" {
		_ = b.tenantId.Set(id)
	}
}

// TenantId implements [RequestContext].
func (b *BaseContext) TenantId() string {
	if id, ok := b.tenantId.Get(); ok {
		return id
	}

	return b.defaultTenantId
}

func (b *BaseContext) Reset() {
	if b == nil {
		return
	}

	b.requestId.Reset()
	b.tenantId.Reset()
	b.resourceId.Reset()
}
