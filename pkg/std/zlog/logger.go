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
	"github.com/sentinez/sentinez/pkg/std/zlog/internal"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
)

var (
	once   sync.Once
	onceLL sync.Once
)

var (
	logcore  = NewLogger()
	logLevel = internal.LevelDebug
)

var _ Logger = (*internal.Core)(nil)

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

// SystemLog define system logger, include grpclog wrapped
type SystemLog interface{ Logger }

// NewLogger creates a new logger instance.
func NewLogger() Logger {
	logger := newLogger().Sugar()

	return internal.NewCore(logger, logLevel.Int())
}

// NewSystemLog init all system log,
// like logger global variable, grpclog global variable
func NewSystemLog() SystemLog {
	once.Do(func() {
		logger := newLogger().Sugar()
		logcore = internal.NewCore(logger, logLevel.Int())

		grpclog.SetLoggerV2(
			internal.NewCore(logger, internal.LevelWarning.Int()))
	})

	return logcore
}

// SetLogLevel set logger default level
func SetLogLevel(ll string) {
	onceLL.Do(func() {
		logLevel = getLogLevel(ll)
	})
}

func getLogLevel(logLevel string) internal.Level {
	level := internal.LevelDebug
	switch logLevel {
	case "debug":
		level = internal.LevelDebug
	case "info":
		level = internal.LevelInfo
	case "warn":
		level = internal.LevelWarning
	case "error":
		level = internal.LevelError
	default:
		level = internal.LevelDebug
	}

	return level
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
			enc.AppendString(
				fmt.Sprintf("%s %s",
					color.Green.Add("[SENTINEZ]"), t.Format(time.DateTime)))
		},
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	// Console output
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	core := zapcore.NewCore(
		consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))
	return logger
}

// Info logs an info message.
func Info(message ...any) {
	logcore.Info(message...)
}

// Infof logs an info message with a format.
func Infof(template string, message ...any) {
	logcore.Infof(template, message...)
}

// Debug logs a debug message.
func Debug(message ...any) {
	logcore.Debug(message...)
}

// Debugf logs a debug message.
func Debugf(template string, message ...any) {
	logcore.Debugf(template, message...)
}

// Error logs an error message.
func Error(message ...any) {
	logcore.Error(message...)
}

// Errorf logs an error message with a format.
func Errorf(template string, message ...any) {
	logcore.Errorf(template, message...)
}

// Warn logs an warn message.
func Warn(message ...any) {
	logcore.Warning(message...)
}

// Warnf logs an error message with a format.
func Warnf(template string, message ...any) {
	logcore.Warningf(template, message...)
}

// Fatal logs a fatal message.
func Fatal(message ...any) {
	logcore.Fatal(message...)
}

// Fatalf logs a fatal message.
func Fatalf(template string, message ...any) {
	logcore.Fatalf(template, message...)
}
