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

// Package internal provides the logger for the package.
package internal

import (
	"github.com/sentinez/sentinez/pkg/common/color"
	"go.uber.org/zap"
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

var header = color.White.Add(">>> ")

// Core is the logger for the package.
type Core struct {
	Verbosity int
	Logger    *zap.SugaredLogger
}

func (c *Core) addPrefix(args ...any) []any {
	return append([]any{header}, args...)
}

func (c *Core) addPrefixFormat(format string) string {
	return header + format
}

// Debug logs a debug message.
func (c *Core) Debug(args ...any) {
	if c.V(LevelDebug.Int()) {
		c.Logger.Debug(c.addPrefix(args...)...)
	}
}

// Debugln logs a debug message.
func (c *Core) Debugln(args ...any) { c.Debug(args...) }

// Debugf logs a debug message with a format.
func (c *Core) Debugf(format string, args ...any) {
	if c.V(LevelDebug.Int()) {
		c.Logger.Debugf(c.addPrefixFormat(format), args...)
	}
}

// Info logs an info message.
func (c *Core) Info(args ...any) {
	if c.V(LevelInfo.Int()) {
		c.Logger.Info(c.addPrefix(args...)...)
	}
}

// Infoln logs an info message.
func (c *Core) Infoln(args ...any) {
	c.Info(args...)
}

// Infof logs an info message with a format.
func (c *Core) Infof(format string, args ...any) {
	if c.V(LevelInfo.Int()) {
		c.Logger.Infof(c.addPrefixFormat(format), args...)
	}
}

// Warning logs a warning message.
func (c *Core) Warning(args ...any) {
	if c.V(LevelWarning.Int()) {
		c.Logger.Warn(c.addPrefix(args...)...)
	}
}

// Warningln logs a warning message.
func (c *Core) Warningln(args ...any) {
	c.Warning(args...)
}

// Warningf logs a warning message with a format.s
func (c *Core) Warningf(format string, args ...any) {
	if c.V(LevelWarning.Int()) {
		c.Logger.Warnf(c.addPrefixFormat(format), args...)
	}
}

// Error logs an error message.
func (c *Core) Error(args ...any) {
	if c.V(LevelError.Int()) {
		c.Logger.Error(c.addPrefix(args...)...)
	}
}

// Errorln logs an error message.
func (c *Core) Errorln(args ...any) {
	c.Error(args...)
}

// Errorf logs an error message with a format.
func (c *Core) Errorf(format string, args ...any) {
	if c.V(LevelError.Int()) {
		c.Logger.Errorf(c.addPrefixFormat(format), args...)
	}
}

// Fatal logs a fatal message.
func (c *Core) Fatal(args ...any) {
	if c.V(LevelFatal.Int()) {
		c.Logger.Fatal(c.addPrefix(args...)...)
	}
}

// Fatalln logs a fatal message.
func (c *Core) Fatalln(args ...any) { c.Fatal(args...) }

// Fatalf logs a fatal message with a format.
func (c *Core) Fatalf(format string, args ...any) {
	if c.V(LevelFatal.Int()) {
		c.Logger.Fatalf(c.addPrefixFormat(format), args...)
	}
}

// V reports whether verbosity level l is at least the requested verbose level.
func (c *Core) V(l int) bool {
	return l >= c.Verbosity
}

// Sync flushes the log.
func (c *Core) Sync() error {
	return c.Logger.Sync()
}

// NewCore creates a new Core.
func NewCore(logger *zap.SugaredLogger, verbosity int) *Core {
	return &Core{Logger: logger, Verbosity: verbosity}
}
