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
	"io"

	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	loggerEvent = "event"
	loggerKind  = "kind"
)

var (
	_ Log       = (*logger)(nil)
	_ LogCloser = (*logCloser)(nil)
)

type Log interface {
	Info(msg string, event proto.Message)
	Debug(msg string, event proto.Message)
	Warn(msg string, event proto.Message)
	Error(msg string, event proto.Message)
	V(l int) bool
	Sync() error
}

type LogCloser interface {
	Info(msg string, event proto.Message, closer io.Closer)
	Debug(msg string, event proto.Message, closer io.Closer)
	Warn(msg string, event proto.Message, closer io.Closer)
	Error(msg string, event proto.Message, closer io.Closer)
	V(l int) bool
	Sync() error
}

func NewLog(named string, logKind typepb.LogKind, level Level) Log {
	logger := configJSONLogger(named)
	return createLogger(logger, logKind, ToLevel(level.String()).Int())
}

func NewLogCloser(named string, logKind typepb.LogKind, level Level) LogCloser {
	logger := configJSONLogger(named)
	return &logCloser{
		logger: createLogger(logger, logKind, ToLevel(level.String()).Int()),
	}

}

func createLogger(log *zap.Logger,
	kind typepb.LogKind, verbosity int) *logger {
	return &logger{log: log, verbosity: verbosity, kind: kind}
}

type logger struct {
	log       *zap.Logger
	kind      typepb.LogKind
	verbosity int
}

// Debug implements Logger.
func (l *logger) Debug(msg string, event proto.Message) {
	if l.V(LevelDebug.Int()) {
		l.log.Debug(msg,
			zap.Object(loggerEvent, marshaler(event)))
	}
}

// Error implements Logger.
func (l *logger) Error(msg string, event proto.Message) {
	if l.V(LevelError.Int()) {
		l.log.Error(msg,
			zap.Object(loggerEvent, marshaler(event)))
	}
}

// Info implements Logger.
func (l *logger) Info(msg string, event proto.Message) {
	if l.V(LevelInfo.Int()) {
		l.log.Info(msg, zap.String(loggerKind, l.kind.String()),
			zap.Object(loggerEvent, marshaler(event)))
	}
}

// Sync implements Logger.
func (l *logger) Sync() error {
	return l.log.Sync()
}

// V implements Logger.
func (l *logger) V(ll int) bool {
	return ll >= l.verbosity
}

// Warn implements Logger.
func (l *logger) Warn(msg string, event proto.Message) {
	if l.V(LevelWarning.Int()) {
		l.log.Warn(msg,
			zap.Object(loggerEvent, marshaler(event)))
	}
}

type logCloser struct {
	*logger
}

// Debug implements [LogCloser].
// Subtle: this method shadows the method (*logger).Debug of logCloser.logger.
func (l *logCloser) Debug(msg string, event proto.Message, closer io.Closer) {
	l.logger.Debug(msg, event)
	_ = closer.Close()
}

// Error implements [LogCloser].
// Subtle: this method shadows the method (*logger).Error of logCloser.logger.
func (l *logCloser) Error(msg string, event proto.Message, closer io.Closer) {
	l.logger.Error(msg, event)
	_ = closer.Close()
}

// Info implements [LogCloser].
// Subtle: this method shadows the method (*logger).Info of logCloser.logger.
func (l *logCloser) Info(msg string, event proto.Message, closer io.Closer) {
	l.logger.Info(msg, event)
	_ = closer.Close()
}

// Sync implements [LogCloser].
// Subtle: this method shadows the method (*logger).Sync of logCloser.logger.
func (l *logCloser) Sync() error {
	return l.logger.Sync()
}

// V implements [LogCloser].
// Subtle: this method shadows the method (*logger).V of logCloser.logger.
func (l *logCloser) V(level int) bool {
	return l.logger.V(level)
}

// Warn implements [LogCloser].
// Subtle: this method shadows the method (*logger).Warn of logCloser.logger.
func (l *logCloser) Warn(msg string, event proto.Message, closer io.Closer) {
	l.logger.Warn(msg, event)
	_ = closer.Close()
}
