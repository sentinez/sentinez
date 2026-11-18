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
	"github.com/sentinez/controlplane/pkg/auditchange"
	auditpb "github.com/sentinez/sentinez/api/proto/sentinez/types/audit/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// changeOf returns the fields of the update recorded on c whose value it
// changed, their values as redact writes them; nil when no update was
// recorded or it changed nothing.
func changeOf(c *auditchange.Change) *auditpb.Change {
	before, after, paths := c.Get()
	if before == nil || after == nil {
		return nil
	}

	old, now := before.ProtoReflect(), after.ProtoReflect()
	if old.Descriptor() != now.Descriptor() {
		return nil
	}

	// the raw values are compared: a redacted field shows as changed
	changed := changedFields(old, now, paths)
	if len(changed) == 0 {
		return nil
	}

	// only the changed fields are redacted and rendered
	shownOld := redact(only(old, changed))
	shownNow := redact(only(now, changed))
	fields := make([]*auditpb.FieldChange, 0, len(changed))
	for _, fd := range changed {
		fields = append(fields, &auditpb.FieldChange{
			Field:  string(fd.Name()),
			Before: shownOld.GetFields()[fd.JSONName()],
			After:  shownNow.GetFields()[fd.JSONName()],
		})
	}

	return &auditpb.Change{Fields: fields}
}

// changedFields returns the fields named in paths that differ between
// before and after, two messages of the same type, in the order of paths.
func changedFields(before, after protoreflect.Message,
	paths []string) []protoreflect.FieldDescriptor {
	var changed []protoreflect.FieldDescriptor
	for _, path := range paths {
		fd := after.Descriptor().Fields().ByName(protoreflect.Name(path))
		if fd != nil && changedField(before, after, fd) {
			changed = append(changed, fd)
		}
	}

	return changed
}

// only returns a new message of the type of m holding only its fields fds.
func only(m protoreflect.Message,
	fds []protoreflect.FieldDescriptor) proto.Message {
	out := m.New()
	for _, fd := range fds {
		if m.Has(fd) {
			out.Set(fd, m.Get(fd))
		}
	}

	return out.Interface()
}

// changedField reports whether fd differs between before and after, two
// messages of the same type.
func changedField(before, after protoreflect.Message,
	fd protoreflect.FieldDescriptor) bool {
	if before.Has(fd) != after.Has(fd) {
		return true
	}

	return !before.Get(fd).Equal(after.Get(fd))
}
