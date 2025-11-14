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

// Package logger provides the logger for the service.
package zlog

import (
	"testing"
)

func TestInfo(_ *testing.T) {
	Info("This is an info message")
}

func TestInfof(_ *testing.T) {
	Infof("This is an info message with format: %s", "formatted string")
}

func TestDebug(_ *testing.T) {
	Debug("This is a debug message")
}

func TestError(_ *testing.T) {
	Error("This is an error message")
}

func TestErrorf(_ *testing.T) {
	Errorf("This is an error message with format: %s", "formatted string")
}

func TestFatal(_ *testing.T) {
	// Note: Fatal will call os.Exit(1) after logging the message, which will
	// terminate the test.To test Fatal, you might need to mock os.Exit or
	// run it in a separate process. Here, we just demonstrate the call.
	// Fatal("This is a fatal message")
}
