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

// Package api is the Vercel serverless entrypoint. vercel.json rewrites
// every path to /api/index, and the original path is kept in r.URL.
package api

import (
	"net/http"

	"github.com/sentinez/go-import/vanity"
)

// Handler is invoked by the Vercel Go runtime.
func Handler(w http.ResponseWriter, r *http.Request) {
	vanity.New(vanity.DefaultConfig()).ServeHTTP(w, r)
}
