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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/structpb"
)

const _redacted = "REDACTED"

// redact returns req as a struct, with the fields marked debug_redact
// hidden: a string or bytes field reads "REDACTED", any other is cleared.
// req itself is left as it is. It returns nil for a nil or unmarshalable
// request, the entry then has no request.
func redact(req proto.Message) *structpb.Struct {
	if req == nil {
		return nil
	}

	clone := proto.Clone(req)
	hide(clone.ProtoReflect())

	raw, err := protojson.Marshal(clone)
	if err != nil {
		return nil
	}

	out := &structpb.Struct{}
	if err := protojson.Unmarshal(raw, out); err != nil {
		return nil
	}

	return out
}

// hide hides the debug_redact fields of m and of the messages nested in it.
func hide(m protoreflect.Message) {
	fields := m.Descriptor().Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		if !m.Has(fd) {
			continue
		}

		if isRedacted(fd) {
			hideField(m, fd)
			continue
		}

		if fd.Kind() == protoreflect.MessageKind {
			hideNested(m.Get(fd), fd)
		}
	}
}

func isRedacted(fd protoreflect.FieldDescriptor) bool {
	opts, _ := fd.Options().(*descriptorpb.FieldOptions)

	return opts.GetDebugRedact()
}

func hideField(m protoreflect.Message, fd protoreflect.FieldDescriptor) {
	if fd.IsList() || fd.IsMap() {
		m.Clear(fd)
		return
	}

	switch fd.Kind() {
	case protoreflect.StringKind:
		m.Set(fd, protoreflect.ValueOfString(_redacted))
	case protoreflect.BytesKind:
		m.Set(fd, protoreflect.ValueOfBytes([]byte(_redacted)))
	default:
		m.Clear(fd)
	}
}

func hideNested(v protoreflect.Value, fd protoreflect.FieldDescriptor) {
	switch {
	case fd.IsList():
		list := v.List()
		for i := range list.Len() {
			hide(list.Get(i).Message())
		}
	case fd.IsMap():
		if fd.MapValue().Kind() != protoreflect.MessageKind {
			return
		}

		v.Map().Range(func(_ protoreflect.MapKey,
			mv protoreflect.Value) bool {
			hide(mv.Message())
			return true
		})
	default:
		hide(v.Message())
	}
}
