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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const _console = "https://console.example.com"

// serve sends a request with the given method and origin through CORS.
func serve(origins []string, method, origin string, preflight bool,
) *httptest.ResponseRecorder {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	r := httptest.NewRequest(method, "/iam/login", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if preflight {
		r.Header.Set("Access-Control-Request-Method", "PUT")
	}

	w := httptest.NewRecorder()
	CORS(origins)(next).ServeHTTP(w, r)

	return w
}

func TestCORS_AllowedOrigin(t *testing.T) {
	w := serve([]string{_console}, http.MethodGet, _console, false)

	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.Equal(t, _console, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true",
		w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Values("Vary"), "Origin")
}

func TestCORS_Preflight(t *testing.T) {
	w := serve([]string{_console}, http.MethodOptions, _console, true)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, _console, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"),
		"Authorization")
}

func TestCORS_UnlistedOrigin(t *testing.T) {
	tests := []struct {
		name         string
		giveOrigins  []string
		giveOrigin   string
		givePreflght bool
		wantStatus   int
	}{
		{"another site", []string{_console}, "https://evil.example.com",
			false, http.StatusTeapot},
		{"a sub domain", []string{_console}, "https://x.console.example.com",
			false, http.StatusTeapot},
		{"another scheme", []string{_console}, "http://console.example.com",
			false, http.StatusTeapot},
		{"its preflight", []string{_console}, "https://evil.example.com",
			true, http.StatusTeapot},
		{"no origins listed", nil, _console, false, http.StatusTeapot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodGet
			if tt.givePreflght {
				method = http.MethodOptions
			}

			w := serve(tt.giveOrigins, method, tt.giveOrigin, tt.givePreflght)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
			assert.Contains(t, w.Header().Values("Vary"), "Origin")
		})
	}
}

func TestCORS_SameOriginCall(t *testing.T) {
	w := serve([]string{_console}, http.MethodGet, "", false)

	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Values("Vary"))
}
