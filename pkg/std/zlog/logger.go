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

// Package stdlog provides the logger for the service.
package zlog

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sentinez/sentinez/pkg/common/color"
	"github.com/sentinez/sentinez/pkg/std/version"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once sync.Once
)

var (
	_    Logger = (*zcore)(nil)
	core        = NewConsole(LevelDebug)
)

// Logger define default logger for logger
type Logger interface {
	Info(args ...any)
	Infof(template string, args ...any)
	Infoln(args ...any)

	Debug(args ...any)
	Debugf(template string, args ...any)
	Debugln(args ...any)

	Warning(args ...any)
	Warningf(template string, args ...any)
	Warningln(args ...any)

	Error(args ...any)
	Errorf(template string, args ...any)
	Errorln(args ...any)

	Fatal(args ...any)
	Fatalf(template string, args ...any)
	Fatalln(args ...any)

	V(l int) bool
	Sync() error
}

// newLogger creates a new logger.
func newLogger() *zap.Logger {
	config := zapcore.EncoderConfig{
		TimeKey:       "timestamp",
		LevelKey:      "level",
		NameKey:       "logger",
		MessageKey:    "message",
		StacktraceKey: "stacktrace",
		CallerKey:     "caller",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(time.DateTime))
		},
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	// Console output
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	core := zapcore.NewCore(
		consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCaller(),
		zap.AddCallerSkip(2), zap.AddStacktrace(zapcore.FatalLevel))

	logger = logger.Named(color.Green.Add(fmt.Sprintf("[%s]", version.Code)))
	return logger
}

func NewConsole(level Level) Logger {
	logger := newLogger().Sugar()
	return createZCore(logger, ToLevel(level.String()).Int())
}

// SetLogLevel set logger default level
func SetLogLevel(ll string) {
	once.Do(func() {
		level := ToLevel(ll)
		core = NewConsole(level)
	})
}

// Info logs an info message.
func Info(message ...any) {
	core.Info(message...)
}

// Infof logs an info message with a format.
func Infof(template string, message ...any) {
	core.Infof(template, message...)
}

// Debug logs a debug message.
func Debug(message ...any) {
	core.Debug(message...)
}

// Debugf logs a debug message.
func Debugf(template string, message ...any) {
	core.Debugf(template, message...)
}

// Error logs an error message.
func Error(message ...any) {
	core.Error(message...)
}

// Errorf logs an error message with a format.
func Errorf(template string, message ...any) {
	core.Errorf(template, message...)
}

// Warn logs an warn message.
func Warn(message ...any) {
	core.Warning(message...)
}

// Warnf logs an error message with a format.
func Warnf(template string, message ...any) {
	core.Warningf(template, message...)
}

// Fatal logs a fatal message.
func Fatal(message ...any) {
	core.Fatal(message...)
}

// Fatalf logs a fatal message.
func Fatalf(template string, message ...any) {
	core.Fatalf(template, message...)
}
