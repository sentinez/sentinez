package zlog

import (
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	loggerEvent = "event"
	loggerKind  = "kind"
)

var _ Logger = (*logger)(nil)

type Logger interface {
	Info(kind string, msg string, event proto.Message)
	Debug(msg string, event proto.Message)
	Warn(msg string, event proto.Message)
	Error(msg string, event proto.Message)
	V(l int) bool
	Sync() error
}

func NewJSON(kind common.Kind, level Level) Logger {
	logger := configJSONLogger(kind.String())
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
func (l *logger) Info(kind string, msg string, event proto.Message) {
	if l.V(LevelInfo.Int()) {
		l.log.Info(msg, zap.String(loggerKind, kind),
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
