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

package szhttp

import (
	"net/http"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/shared/bytesconv"
)

// Cookie returns the value of the named request cookie.
func Cookie(req Request, name string) (string, bool) {
	header := bytesconv.B2s(req.Header(szbytestr.HeaderCookie))
	cookies, err := http.ParseCookie(header)
	if err != nil {
		return "", false
	}

	for _, c := range cookies {
		if c.Name == name {
			return c.Value, true
		}
	}

	return "", false
}
