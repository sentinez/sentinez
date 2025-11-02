package zlog

import (
	"sync"

	"go.uber.org/zap"
)

var (
	logger *zap.Logger
	once   sync.Once
)

func console() *zap.SugaredLogger {
	once.Do(func() {
		logger, _ = zap.NewDevelopment(zap.AddCallerSkip(2))
	})
	return logger.Sugar()
}

// Info logs an info message.
func Info(message ...any) {
	console().Info(message...)
}

// Infof logs an info message with a format.
func Infof(template string, message ...any) {
	console().Infof(template, message...)
}

// Debug logs a debug message.
func Debug(message ...any) {
	console().Debug(message...)
}

// Debugf logs a debug message.
func Debugf(template string, message ...any) {
	console().Debugf(template, message...)
}

// Error logs an error message.
func Error(message ...any) {
	console().Error(message...)
}

// Errorf logs an error message with a format.
func Errorf(template string, message ...any) {
	console().Errorf(template, message...)
}

// Warn logs an warn message.
func Warn(message ...any) {
	console().Warn(message...)
}

// Warnf logs an error message with a format.
func Warnf(template string, message ...any) {
	console().Warnf(template, message...)
}

// Fatal logs a fatal message.
func Fatal(message ...any) {
	console().Fatal(message...)
}

// Fatalf logs a fatal message.
func Fatalf(template string, message ...any) {
	console().Fatalf(template, message...)
}
