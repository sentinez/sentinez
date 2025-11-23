// Copyright 2025 Duc-Hung Ho.
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

package zlog

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	config = zapcore.EncoderConfig{
		TimeKey:       "timestamp",
		LevelKey:      "level",
		NameKey:       "name",
		MessageKey:    "message",
		StacktraceKey: "stacktrace",
		CallerKey:     "caller",
		EncodeTime:    zap.NewDevelopmentEncoderConfig().EncodeTime,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}

	onceSetScope sync.Once
	onceSetLevel sync.Once
)

const (
	// LevelDebug is the lowest level of verbosity.
	LevelDebug Level = 0

	// LevelInfo is the default level of verbosity.
	LevelInfo Level = 1

	// LevelWarning is the default level of verbosity.
	LevelWarning Level = 2

	// LevelError is the default level of verbosity.
	LevelError Level = 3

	// LevelFatal is the highest level of verbosity.
	LevelFatal Level = 4
)

// Level is log level of logger
type Level int

// String return String type
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarning:
		return "warning"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// Int return integer type
func (l Level) Int() int {
	return int(l)
}

// configConsoleLogger creates a new logger.
func configConsoleLogger(scope string) *zap.Logger {
	var consoleConf = config
	consoleConf.EncodeTime = zap.NewDevelopmentEncoderConfig().EncodeTime

	// Console output
	consoleEncoder := zapcore.NewConsoleEncoder(consoleConf)

	core := zapcore.NewCore(consoleEncoder,
		zapcore.Lock(os.Stdout), zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCaller(),
		zap.AddCallerSkip(2), zap.AddStacktrace(zapcore.FatalLevel))

	logger = logger.Named(scope)
	return logger
}

func configJSONLogger(scope string) *zap.Logger {
	var jsonConf = config
	jsonConf.EncodeTime = zap.NewProductionEncoderConfig().EncodeTime
	jsonConf.EncodeLevel = zapcore.CapitalLevelEncoder

	jsonEncoder := zapcore.NewJSONEncoder(jsonConf)

	core := zapcore.NewCore(jsonEncoder,
		zapcore.Lock(os.Stdout), zapcore.DebugLevel)

	logger := zap.New(core)

	logger = logger.Named(scope)
	return logger
}

// SetLogLevel set logger default level
func SetLogLevel(ll Level) {
	onceSetLevel.Do(func() {
		consoleLevel = ll
		console = NewConsole("", consoleLevel)
	})
}

func SetScopeLogLevel(scope string, ll Level) {
	onceSetScope.Do(func() {
		consoleLevel = ll
		console = NewConsole(scope, consoleLevel)
	})
}

func ToLevel(logLevel string) Level {
	level := LevelDebug
	switch logLevel {
	case "debug":
		level = LevelDebug
	case "info":
		level = LevelInfo
	case "warn":
		level = LevelWarning
	case "error":
		level = LevelError
	case "fatal":
		level = LevelFatal
	default:
		level = LevelDebug
	}

	return level
}

func marshaler(event proto.Message) zapcore.ObjectMarshaler {
	return zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		return marshalProto(enc, event)
	})
}

func marshalProto(enc zapcore.ObjectEncoder, msg proto.Message) error {
	m := msg.ProtoReflect()
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		name := string(fd.Name())

		switch {
		case fd.IsList():
			// repeated field -> log slice
			arr := make([]any, 0, v.List().Len())
			for i := 0; i < v.List().Len(); i++ {
				arr = append(arr, valueOf(fd, v.List().Get(i)))
			}
			_ = enc.AddReflected(name, arr)

		case fd.IsMap():
			// map field -> log map
			tmp := make(map[string]any)
			v.Map().Range(
				func(k protoreflect.MapKey, v protoreflect.Value) bool {
					tmp[k.String()] = valueOf(fd.MapValue(), v)
					return true
				})
			_ = enc.AddReflected(name, tmp)

		default:
			// scalar field
			addScalar(enc, name, fd, v)
		}
		return true
	})
	return nil
}

// nolint:funlen
func addScalar(enc zapcore.ObjectEncoder, key string,
	fd protoreflect.FieldDescriptor, v protoreflect.Value) {
	switch fd.Kind() {
	case protoreflect.StringKind:
		enc.AddString(key, v.String())
	case protoreflect.Int32Kind, protoreflect.Sint32Kind,
		protoreflect.Sfixed32Kind:
		enc.AddInt32(key, int32(v.Int()))
	case protoreflect.Int64Kind, protoreflect.Sint64Kind,
		protoreflect.Sfixed64Kind:
		enc.AddInt64(key, v.Int())
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		enc.AddUint32(key, uint32(v.Uint()))
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		enc.AddUint64(key, v.Uint())
	case protoreflect.FloatKind:
		enc.AddFloat32(key, float32(v.Float()))
	case protoreflect.DoubleKind:
		enc.AddFloat64(key, v.Float())
	case protoreflect.BoolKind:
		enc.AddBool(key, v.Bool())
	case protoreflect.BytesKind:
		enc.AddBinary(key, v.Bytes())
	case protoreflect.EnumKind:
		enc.AddInt32(key, int32(v.Enum()))
	case protoreflect.MessageKind:
		_ = enc.AddObject(key, zapcore.ObjectMarshalerFunc(
			func(inner zapcore.ObjectEncoder) error {
				return marshalProto(inner, v.Message().Interface())
			}))
	default:
		_ = enc.AddReflected(key, v.Interface())
	}
}

// nolint:funlen
func valueOf(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	switch fd.Kind() {
	case protoreflect.StringKind:
		return v.String()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind,
		protoreflect.Sfixed32Kind:
		return int32(v.Int())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind,
		protoreflect.Sfixed64Kind:
		return v.Int()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return uint32(v.Uint())
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return v.Uint()
	case protoreflect.FloatKind:
		return float32(v.Float())
	case protoreflect.DoubleKind:
		return v.Float()
	case protoreflect.BoolKind:
		return v.Bool()
	case protoreflect.BytesKind:
		return v.Bytes()
	case protoreflect.EnumKind:
		return int32(v.Enum())
	case protoreflect.MessageKind:

		m := make(map[string]any)
		v.Message().Range(
			func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
				m[string(fd.Name())] = valueOf(fd, val)
				return true
			})
		return m
	default:
		return nil
	}
}
