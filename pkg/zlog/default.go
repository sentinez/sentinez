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

var (
	console = NewDefaultConsole(LevelDebug)
)

// Info logs an info message.
func Info(message ...any) {
	console.Info(message...)
}

// Infof logs an info message with a format.
func Infof(template string, message ...any) {
	console.Infof(template, message...)
}

// Debug logs a debug message.
func Debug(message ...any) {
	console.Debug(message...)
}

// Debugf logs a debug message.
func Debugf(template string, message ...any) {
	console.Debugf(template, message...)
}

// Error logs an error message.
func Error(message ...any) {
	console.Error(message...)
}

// Errorf logs an error message with a format.
func Errorf(template string, message ...any) {
	console.Errorf(template, message...)
}

// Warn logs an warn message.
func Warn(message ...any) {
	console.Warning(message...)
}

// Warnf logs an error message with a format.
func Warnf(template string, message ...any) {
	console.Warningf(template, message...)
}

// Fatal logs a fatal message.
func Fatal(message ...any) {
	console.Fatal(message...)
}

// Fatalf logs a fatal message.
func Fatalf(template string, message ...any) {
	console.Fatalf(template, message...)
}
