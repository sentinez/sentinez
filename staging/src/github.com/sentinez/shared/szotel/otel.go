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

// Package szotel decodes the OTLP log records szlog exports (as read back
// from the collector's Kafka topics) into the proto events szlog was given.
package szotel

import (
	"errors"
	"fmt"
	"strconv"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// attribute keys szlog writes; keep in sync with szlog/logging.go
const (
	_eventKey = "event"
	_kindKey  = "kind"
)

// ErrNoEvent is returned by Unmarshal when the record carries no event.
var ErrNoEvent = errors.New("szotel: record has no event")

// Kind returns the log type szlog stamped on rec, or LOG_TYPE_UNSPECIFIED
// when it has none (szlog only stamps Info records).
func Kind(rec *logspb.LogRecord) typepb.LogType {
	v := attributeOf(rec.GetAttributes(), _kindKey)
	return typepb.LogType(typepb.LogType_value[v.GetStringValue()])
}

// Unmarshal decodes the event szlog attached to rec into msg, the inverse
// of szlog's proto-to-zap marshaling as converted by otelzap. msg is reset
// first; attributes naming no field of msg are ignored.
func Unmarshal(rec *logspb.LogRecord, msg proto.Message) error {
	v := attributeOf(rec.GetAttributes(), _eventKey)
	if v == nil {
		return ErrNoEvent
	}

	kvs, err := kvlistOf(v)
	if err != nil {
		return fmt.Errorf("szotel: event: %w", err)
	}

	proto.Reset(msg)
	if err := decodeMessage(msg.ProtoReflect(), kvs); err != nil {
		return fmt.Errorf("szotel: %w", err)
	}

	return nil
}

func attributeOf(attrs []*commonpb.KeyValue, key string) *commonpb.AnyValue {
	for _, kv := range attrs {
		if kv.GetKey() == key {
			return kv.GetValue()
		}
	}

	return nil
}

func decodeMessage(m protoreflect.Message, kvs []*commonpb.KeyValue) error {
	fields := m.Descriptor().Fields()
	for _, kv := range kvs {
		fd := fields.ByName(protoreflect.Name(kv.GetKey()))
		if fd == nil {
			continue
		}
		if err := decodeField(m, fd, kv.GetValue()); err != nil {
			return fmt.Errorf("%s: %w", fd.Name(), err)
		}
	}

	return nil
}

func decodeField(m protoreflect.Message, fd protoreflect.FieldDescriptor,
	v *commonpb.AnyValue) error {
	switch {
	case fd.IsList():
		return decodeList(m.Mutable(fd).List(), fd, v)
	case fd.IsMap():
		return decodeMap(m.Mutable(fd).Map(), fd, v)
	case fd.Message() != nil:
		kvs, err := kvlistOf(v)
		if err != nil {
			return err
		}
		return decodeMessage(m.Mutable(fd).Message(), kvs)
	}

	val, err := scalarOf(fd, v)
	if err != nil {
		return err
	}
	m.Set(fd, val)

	return nil
}

func decodeList(l protoreflect.List, fd protoreflect.FieldDescriptor,
	v *commonpb.AnyValue) error {
	arr, ok := v.GetValue().(*commonpb.AnyValue_ArrayValue)
	if !ok {
		return mismatch("array", v)
	}

	for i, elem := range arr.ArrayValue.GetValues() {
		val, err := elementOf(l.NewElement, fd, elem)
		if err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		l.Append(val)
	}

	return nil
}

func decodeMap(mp protoreflect.Map, fd protoreflect.FieldDescriptor,
	v *commonpb.AnyValue) error {
	kvs, err := kvlistOf(v)
	if err != nil {
		return err
	}

	for _, kv := range kvs {
		key, err := mapKeyOf(fd.MapKey(), kv.GetKey())
		if err != nil {
			return fmt.Errorf("key %q: %w", kv.GetKey(), err)
		}
		val, err := elementOf(mp.NewValue, fd.MapValue(), kv.GetValue())
		if err != nil {
			return fmt.Errorf("[%q]: %w", kv.GetKey(), err)
		}
		mp.Set(key, val)
	}

	return nil
}

// elementOf decodes one list element or map value; newMessage allocates it
// when fd is a message field.
func elementOf(newMessage func() protoreflect.Value,
	fd protoreflect.FieldDescriptor,
	v *commonpb.AnyValue) (protoreflect.Value, error) {
	if fd.Message() == nil {
		return scalarOf(fd, v)
	}

	kvs, err := kvlistOf(v)
	if err != nil {
		return protoreflect.Value{}, err
	}
	val := newMessage()

	return val, decodeMessage(val.Message(), kvs)
}

// mapKeyOf parses a map key otelzap formatted with "%+v".
func mapKeyOf(fd protoreflect.FieldDescriptor,
	s string) (protoreflect.MapKey, error) {
	var val protoreflect.Value
	var err error
	switch fd.Kind() {
	case protoreflect.StringKind:
		val = protoreflect.ValueOfString(s)
	case protoreflect.BoolKind:
		var b bool
		b, err = strconv.ParseBool(s)
		val = protoreflect.ValueOfBool(b)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind,
		protoreflect.Sfixed32Kind:
		var n int64
		n, err = strconv.ParseInt(s, 10, 32)
		val = protoreflect.ValueOfInt32(int32(n))
	case protoreflect.Int64Kind, protoreflect.Sint64Kind,
		protoreflect.Sfixed64Kind:
		var n int64
		n, err = strconv.ParseInt(s, 10, 64)
		val = protoreflect.ValueOfInt64(n)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		var n uint64
		n, err = strconv.ParseUint(s, 10, 32)
		val = protoreflect.ValueOfUint32(uint32(n))
	default:
		var n uint64
		n, err = strconv.ParseUint(s, 10, 64)
		val = protoreflect.ValueOfUint64(n)
	}

	return val.MapKey(), err
}

// scalarOf decodes a non-message value. szlog writes every integer kind
// and enums as int; otelzap turns a uint64 above MaxInt64 into a double,
// or a string inside lists and maps.
//
// nolint:funlen
func scalarOf(fd protoreflect.FieldDescriptor,
	v *commonpb.AnyValue) (protoreflect.Value, error) {
	switch fd.Kind() {
	case protoreflect.StringKind:
		s, ok := v.GetValue().(*commonpb.AnyValue_StringValue)
		if !ok {
			return protoreflect.Value{}, mismatch("string", v)
		}
		return protoreflect.ValueOfString(s.StringValue), nil
	case protoreflect.BoolKind:
		b, ok := v.GetValue().(*commonpb.AnyValue_BoolValue)
		if !ok {
			return protoreflect.Value{}, mismatch("bool", v)
		}
		return protoreflect.ValueOfBool(b.BoolValue), nil
	case protoreflect.BytesKind:
		b, ok := v.GetValue().(*commonpb.AnyValue_BytesValue)
		if !ok {
			return protoreflect.Value{}, mismatch("bytes", v)
		}
		return protoreflect.ValueOfBytes(b.BytesValue), nil
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return floatOf(fd, v)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		n, err := uint64Of(v)
		return protoreflect.ValueOfUint64(n), err
	}

	return intOf(fd, v)
}

func floatOf(fd protoreflect.FieldDescriptor,
	v *commonpb.AnyValue) (protoreflect.Value, error) {
	f, ok := v.GetValue().(*commonpb.AnyValue_DoubleValue)
	if !ok {
		return protoreflect.Value{}, mismatch("double", v)
	}
	if fd.Kind() == protoreflect.FloatKind {
		return protoreflect.ValueOfFloat32(float32(f.DoubleValue)), nil
	}

	return protoreflect.ValueOfFloat64(f.DoubleValue), nil
}

func uint64Of(v *commonpb.AnyValue) (uint64, error) {
	switch x := v.GetValue().(type) {
	case *commonpb.AnyValue_IntValue:
		return uint64(x.IntValue), nil
	case *commonpb.AnyValue_DoubleValue:
		return uint64(x.DoubleValue), nil
	case *commonpb.AnyValue_StringValue:
		return strconv.ParseUint(x.StringValue, 10, 64)
	}

	return 0, mismatch("int", v)
}

func intOf(fd protoreflect.FieldDescriptor,
	v *commonpb.AnyValue) (protoreflect.Value, error) {
	x, ok := v.GetValue().(*commonpb.AnyValue_IntValue)
	if !ok {
		return protoreflect.Value{}, mismatch("int", v)
	}

	n := x.IntValue
	switch fd.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind,
		protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(int32(n)), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind,
		protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(n), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(uint32(n)), nil
	case protoreflect.EnumKind:
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(n)), nil
	}

	return protoreflect.Value{}, fmt.Errorf("unsupported kind %s", fd.Kind())
}

func kvlistOf(v *commonpb.AnyValue) ([]*commonpb.KeyValue, error) {
	kvs, ok := v.GetValue().(*commonpb.AnyValue_KvlistValue)
	if !ok {
		return nil, mismatch("kvlist", v)
	}

	return kvs.KvlistValue.GetValues(), nil
}

func mismatch(want string, v *commonpb.AnyValue) error {
	return fmt.Errorf("want %s, got %T", want, v.GetValue())
}
