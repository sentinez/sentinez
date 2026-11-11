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

// _resourceIds is the field of the requests that list by resource.
const _resourceIds protoreflect.Name = "resource_ids"

// ResourceScoped is a request that lists by resource: its resource_ids
// field tells which ones, none meaning all.
type ResourceScoped interface {
	proto.Message
	GetResourceIds() []string
}

// Scoped returns a copy of req limited to the resources the caller may see.
// It returns false when the caller may see none: the call must then answer
// empty without running, as a request without resource ids matches all.
func Scoped[R ResourceScoped](ctx context.Context, a *Authorizer, req R,
) (R, bool, error) {
	scope, err := a.ResourceIds(ctx, req.GetResourceIds())
	if err != nil || scope.None {
		return req, false, err
	}

	scoped := proto.CloneOf(req)
	setResourceIds(scoped, scope.Ids)

	return scoped, true, nil
}

func setResourceIds(msg proto.Message, ids []string) {
	m := msg.ProtoReflect()
	list := m.Mutable(m.Descriptor().Fields().ByName(_resourceIds)).List()

	list.Truncate(0)
	for _, id := range ids {
		list.Append(protoreflect.ValueOfString(id))
	}
}
