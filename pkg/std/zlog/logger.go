package zlog

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const loggerEvent = "event"

var _ Logger = (*logger)(nil)

type Logger interface {
	Info(msg string, event proto.Message)
	Debug(msg string, event proto.Message)
	Warn(msg string, event proto.Message)
	Error(msg string, event proto.Message)
	V(l int) bool
	Sync() error
}

func NewJSON(scope string, level Level) Logger {
	logger := configJSONLogger(scope)
	return createLogger(logger, ToLevel(level.String()).Int())
}

func createLogger(log *zap.Logger, verbosity int) Logger {
	return &logger{log: log, verbosity: verbosity}
}

type logger struct {
	log       *zap.Logger
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
		l.log.Info(msg,
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
