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

package interceptor

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	_tenantIdField = "tenant_id"
	_maxIdsDepth   = 3
)

// tenantIdOf returns the first non-empty tenant_id field of msgs or of the
// messages nested in them (e.g. their types.v1.Owner), "" when none has
// one.
func tenantIdOf(msgs ...proto.Message) string {
	for _, m := range msgs {
		if m == nil {
			continue
		}
		if id := findTenantId(m.ProtoReflect(), 0); id != "" {
			return id
		}
	}

	return ""
}

func findTenantId(m protoreflect.Message, depth int) string {
	var id string
	m.Range(func(fd protoreflect.FieldDescriptor,
		v protoreflect.Value) bool {

		id = tenantIdIn(fd, v, depth)
		return id == ""
	})

	return id
}

func tenantIdIn(fd protoreflect.FieldDescriptor, v protoreflect.Value,
	depth int) string {

	switch {
	case fd.IsMap():
		return ""
	case fd.Kind() == protoreflect.StringKind && !fd.IsList():
		if fd.Name() == _tenantIdField {
			return v.String()
		}
	case fd.Kind() == protoreflect.MessageKind && depth < _maxIdsDepth:
		return tenantIdInMessages(fd, v, depth+1)
	}

	return ""
}

func tenantIdInMessages(fd protoreflect.FieldDescriptor,
	v protoreflect.Value, depth int) string {

	if !fd.IsList() {
		return findTenantId(v.Message(), depth)
	}

	list := v.List()
	for i := range list.Len() {
		if id := findTenantId(list.Get(i).Message(), depth); id != "" {
			return id
		}
	}

	return ""
}
