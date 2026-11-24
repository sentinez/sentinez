// Copyright 2026 Duc-Hung Ho.
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

package phase

import "strconv"

// Phase is when a hook runs, relative to its stage.
type Phase int

const (
	// Request (http.request) the hook runs before the stage, and may
	// answer the request: the stage is then skipped.
	Request Phase = iota + 1
	// Response (http.response) the hook runs once the stage, and the
	// ones after it, returned: the response is written, the hook only
	// observes.
	Response
)

func (p Phase) String() string {
	switch p {
	case Request:
		return "http.request"
	case Response:
		return "http.response"
	}

	return "Phase(" + strconv.Itoa(int(p)) + ")"
}
