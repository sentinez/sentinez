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

// Package protoaudit finds the fields x_field.audit marks as the owner of
// an audited call: the audit interceptor reads them, protoc-gen-go-senz
// checks every audited method has one.
package protoaudit

import (
	"slices"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// _maxOwnerDepth how deep in nested messages the marked fields are found.
const _maxOwnerDepth = 3

// OwnerPath a string field x_field.audit marks, last of the fields of its
// path, the messages on the way first.
type OwnerPath struct {
	Audit  *typepb.XAudit
	Fields []protoreflect.FieldDescriptor
}

// OwnerPaths returns the paths of the singular string fields
// x_field.audit marks as a tenant or a resource id in md, in the order of
// its fields, up to _maxOwnerDepth nested messages; maps are skipped.
func OwnerPaths(md protoreflect.MessageDescriptor) []OwnerPath {
	var paths []OwnerPath
	collect(md, nil, &paths)

	return paths
}

func collect(md protoreflect.MessageDescriptor,
	prefix []protoreflect.FieldDescriptor, paths *[]OwnerPath) {
	fields := md.Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		path := append(slices.Clip(prefix), fd)
		switch {
		case fd.IsMap():
		case fd.Kind() == protoreflect.StringKind && !fd.IsList():
			if audit := auditOf(fd); marksOwner(audit) {
				*paths = append(*paths,
					OwnerPath{Audit: audit, Fields: path})
			}
		case fd.Kind() == protoreflect.MessageKind &&
			len(prefix) < _maxOwnerDepth:
			collect(fd.Message(), path, paths)
		}
	}
}

// auditOf returns the x_field.audit of fd, nil when unset.
func auditOf(fd protoreflect.FieldDescriptor) *typepb.XAudit {
	xfield, _ := proto.GetExtension(fd.Options(),
		typepb.E_XField).(*typepb.XField)

	return xfield.GetAudit()
}

// marksOwner reports whether audit marks a tenant or a resource id.
func marksOwner(audit *typepb.XAudit) bool {
	return audit.GetTenantId() || audit.GetResourceId()
}
