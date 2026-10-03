// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracer

import (
	"errors"
	"fmt"
	"time"
)

// StageError reports the chain stage where an error originated.
type StageError struct {
	Stage   string        `json:"stage"`
	Elapsed time.Duration `json:"elapsed"`
	Err     error         `json:"-"`
}

func (e *StageError) Error() string {
	return fmt.Sprintf("%s: %v", e.Stage, e.Err)
}

func (e *StageError) Unwrap() error { return e.Err }

// Origin returns the stage that produced err, if known.
func Origin(err error) (string, bool) {
	if se, ok := errors.AsType[*StageError](err); ok {
		return se.Stage, true
	}

	return "", false
}
