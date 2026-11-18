// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package interceptor

import (
	"github.com/sentinez/controlplane/pkg/auditowner"
	"github.com/sentinez/pkg/szprotobuf/protoaudit"
	"github.com/sentinez/pkg/szsync"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// _ownerPaths the owner paths of each message, by its descriptor.
var _ownerPaths szsync.Map[protoreflect.MessageDescriptor,
	[]protoaudit.OwnerPath]

// setOwner sets the unset ids of o to the first non-empty field
// x_field.audit marks of each in msgs, its request then its response.
func setOwner(o *auditowner.Owner, msgs ...proto.Message) {
	for _, m := range msgs {
		if m == nil {
			continue
		}
		r := m.ProtoReflect()
		for _, path := range ownerPathsOf(r.Descriptor()) {
			if tenantId, resourceId := o.Ids(); tenantId != "" &&
				resourceId != "" {
				return
			}
			id := stringAt(r, path.Fields)
			o.Set(idIf(path.Audit.GetTenantId(), id),
				idIf(path.Audit.GetResourceId(), id))
		}
	}
}

// idIf returns id when marked, else "".
func idIf(marked bool, id string) string {
	if !marked {
		return ""
	}

	return id
}

// stringAt returns the string field at the end of fields of m, "" when
// unset; the first element of a repeated message on the way.
func stringAt(m protoreflect.Message,
	fields []protoreflect.FieldDescriptor) string {
	last := len(fields) - 1
	for _, fd := range fields[:last] {
		v := m.Get(fd)
		if !fd.IsList() {
			m = v.Message()
			continue
		}
		if v.List().Len() == 0 {
			return ""
		}
		m = v.List().Get(0).Message()
	}

	return m.Get(fields[last]).String()
}

// ownerPathsOf returns the protoaudit.OwnerPaths of md, cached.
func ownerPathsOf(md protoreflect.MessageDescriptor) []protoaudit.OwnerPath {
	if paths, ok := _ownerPaths.Load(md); ok {
		return paths
	}

	paths := protoaudit.OwnerPaths(md)
	_ownerPaths.Store(md, paths)

	return paths
}
