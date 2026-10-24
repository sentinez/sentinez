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

package szlog

import (
	"io"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	_loggingEvent = "event"
	_loggingKind  = "kind"
)

var (
	_ Log       = (*logging)(nil)
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

func NewLog(named string, logKind typepb.LogType, level Level,
	opts ...Option) Log {

	o := newLoggingOptions(opts)
	logging := configJSONLogging(named, o)
	return createLogging(logging, logKind, ToLevel(level.String()).Int(), o)
}

// NewLogCloser creates a LogCloser; pass WithEventExporter to hand its
// events to e.g. a Kafka producer, or WithOTLPProvider to export its
// records through a dedicated OTLP setup.
func NewLogCloser(named string, logKind typepb.LogType, level Level,
	opts ...Option) LogCloser {

	o := newLoggingOptions(opts)
	logging := configJSONLogging(named, o)
	return &logCloser{
		logging: createLogging(logging, logKind,
			ToLevel(level.String()).Int(), o),
	}

}

func createLogging(log *zap.Logger, kind typepb.LogType, verbosity int,
	o loggingOptions) *logging {
	return &logging{log: log, verbosity: verbosity, kind: kind,
		exporter: o.exporter}
}

type logging struct {
	log       *zap.Logger
	kind      typepb.LogType
	verbosity int
	exporter  EventExporter
}

func (l *logging) Debug(msg string, event proto.Message) {
	if l.V(LevelDebug.Int()) {
		l.log.Debug(msg,
			zap.Object(_loggingEvent, marshaler(event)))
		l.export(LevelDebug, msg, event)
	}
}

func (l *logging) Error(msg string, event proto.Message) {
	if l.V(LevelError.Int()) {
		l.log.Error(msg,
			zap.Object(_loggingEvent, marshaler(event)))
		l.export(LevelError, msg, event)
	}
}

func (l *logging) Info(msg string, event proto.Message) {
	if l.V(LevelInfo.Int()) {
		l.log.Info(msg, zap.String(_loggingKind, l.kind.String()),
			zap.Object(_loggingEvent, marshaler(event)))
		l.export(LevelInfo, msg, event)
	}
}

func (l *logging) Sync() error {
	return l.log.Sync()
}

func (l *logging) V(ll int) bool {
	return ll >= l.verbosity
}

func (l *logging) Warn(msg string, event proto.Message) {
	if l.V(LevelWarning.Int()) {
		l.log.Warn(msg,
			zap.Object(_loggingEvent, marshaler(event)))
		l.export(LevelWarning, msg, event)
	}
}

// export hands event to the logger's EventExporter, if it has one.
func (l *logging) export(level Level, msg string, event proto.Message) {
	if l.exporter != nil {
		l.exporter.Export(l.kind, level, msg, event)
	}
}

type logCloser struct {
	*logging
}

func (l *logCloser) Debug(msg string, event proto.Message, closer io.Closer) {
	l.logging.Debug(msg, event)
	_ = closer.Close()
}

func (l *logCloser) Error(msg string, event proto.Message, closer io.Closer) {
	l.logging.Error(msg, event)
	_ = closer.Close()
}

func (l *logCloser) Info(msg string, event proto.Message, closer io.Closer) {
	l.logging.Info(msg, event)
	_ = closer.Close()
}

func (l *logCloser) Sync() error {
	return l.logging.Sync()
}

func (l *logCloser) V(level int) bool {
	return l.logging.V(level)
}

func (l *logCloser) Warn(msg string, event proto.Message, closer io.Closer) {
	l.logging.Warn(msg, event)
	_ = closer.Close()
}
