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
	"github.com/sentinez/shared/color"

	"go.uber.org/zap"
)

var _ Sugar = (*sugard)(nil)

// Sugar define default logger for logger
type Sugar interface {
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

func NewConsole(scope string, level Level) Sugar {
	scope = color.Green.Add(scope)
	log := configConsoleLogger(scope).Sugar()
	return createSugard(log, ToLevel(level.String()).Int())
}

func NewDefaultConsole(level Level) Sugar {
	log := configConsoleLogger("").Sugar()
	return createSugard(log, ToLevel(level.String()).Int())
}

// sugard is the logger for the package.
type sugard struct {
	Verbosity int
	Logger    *zap.SugaredLogger
}

// Debug logs a debug message.
func (c *sugard) Debug(args ...any) {
	if c.V(LevelDebug.Int()) {
		c.Logger.Debug(args...)
	}
}

// Debugln logs a debug message.
func (c *sugard) Debugln(args ...any) { c.Debug(args...) }

// Debugf logs a debug message with a format.
func (c *sugard) Debugf(format string, args ...any) {
	if c.V(LevelDebug.Int()) {
		c.Logger.Debugf(format, args...)
	}
}

// Info logs an info message.
func (c *sugard) Info(args ...any) {
	if c.V(LevelInfo.Int()) {
		c.Logger.Info(args...)
	}
}

// Infoln logs an info message.
func (c *sugard) Infoln(args ...any) {
	c.Info(args...)
}

// Infof logs an info message with a format.
func (c *sugard) Infof(format string, args ...any) {
	if c.V(LevelInfo.Int()) {
		c.Logger.Infof(format, args...)
	}
}

// Warning logs a warning message.
func (c *sugard) Warning(args ...any) {
	if c.V(LevelWarning.Int()) {
		c.Logger.Warn(args...)
	}
}

// Warningln logs a warning message.
func (c *sugard) Warningln(args ...any) {
	c.Warning(args...)
}

// Warningf logs a warning message with a format.s
func (c *sugard) Warningf(format string, args ...any) {
	if c.V(LevelWarning.Int()) {
		c.Logger.Warnf(format, args...)
	}
}

// Error logs an error message.
func (c *sugard) Error(args ...any) {
	if c.V(LevelError.Int()) {
		c.Logger.Error(args...)
	}
}

// Errorln logs an error message.
func (c *sugard) Errorln(args ...any) {
	c.Error(args...)
}

// Errorf logs an error message with a format.
func (c *sugard) Errorf(format string, args ...any) {
	if c.V(LevelError.Int()) {
		c.Logger.Errorf(format, args...)
	}
}

// Fatal logs a fatal message.
func (c *sugard) Fatal(args ...any) {
	if c.V(LevelFatal.Int()) {
		c.Logger.Fatal(args...)
	}
}

// Fatalln logs a fatal message.
func (c *sugard) Fatalln(args ...any) { c.Fatal(args...) }

// Fatalf logs a fatal message with a format.
func (c *sugard) Fatalf(format string, args ...any) {
	if c.V(LevelFatal.Int()) {
		c.Logger.Fatalf(format, args...)
	}
}

// V reports whether verbosity level l is at least the requested verbose level.
func (c *sugard) V(l int) bool {
	return l >= c.Verbosity
}

// Sync flushes the log.
func (c *sugard) Sync() error {
	return c.Logger.Sync()
}

// createSugard creates a new Core.
func createSugard(logger *zap.SugaredLogger, verbosity int) *sugard {
	return &sugard{Logger: logger, Verbosity: verbosity}
}
