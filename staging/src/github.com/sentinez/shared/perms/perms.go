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

package perms

import (
	"errors"
	"slices"

	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
)

// Allow checks if the given console has access to the method.
// The logic follows:
//   - If method is nil, access is allowed.
//   - If method.Ignore is true, access is allowed.
//   - If method.Consoles is empty, access is allowed.
//   - If user's console matches one of the allowed consoles, access is allowed.
//   - Otherwise, access is denied.
func Allow(method *typepb.XMethod, console typepb.Console) error {
	if method == nil {
		return nil
	}

	if method.Ignore {
		return nil
	}

	if len(method.Consoles) == 0 {
		return nil
	}

	if slices.Contains(method.Consoles, console) {
		return nil
	}

	return errors.New("access denied: console not allowed")
}
