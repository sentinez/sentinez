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

package middleware

import (
	"net/http"
	"strings"
)

// Scope restricts m to requests whose path is under one of prefixes, e.g.
// "/iam" covers "/iam" and "/iam/login" but not "/iamx". Every other request
// skips m and goes straight to the next handler.
func Scope(m func(http.Handler) http.Handler,
	prefixes ...string) func(http.Handler) http.Handler {

	return func(h http.Handler) http.Handler {
		scoped := m(h)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if inScope(r.URL.Path, prefixes) {
				scoped.ServeHTTP(w, r)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

func inScope(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		rest, ok := strings.CutPrefix(path, prefix)
		if ok && (rest == "" || rest[0] == '/') {
			return true
		}
	}

	return false
}
