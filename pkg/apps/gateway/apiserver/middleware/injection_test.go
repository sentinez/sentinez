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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	_giveSQLi = `1' OR '1'='1' -- `
	_giveXSS  = `<script>alert(1)</script>`
)

func TestInjection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		giveMethod      string
		giveTarget      string
		giveBody        string
		giveContentType string
		giveHeader      http.Header
		wantCode        int
	}{
		{
			name:       "clean get",
			giveMethod: http.MethodGet,
			giveTarget: "/v1/security/secrules?page=1&search=admin",
			wantCode:   http.StatusOK,
		},
		{
			name:            "clean json body",
			giveMethod:      http.MethodPost,
			giveTarget:      "/v1/iam/login",
			giveBody:        `{"email":"a@b.co","tags":["x","y"],"n":1}`,
			giveContentType: "application/json",
			wantCode:        http.StatusOK,
		},
		{
			name:       "sqli in path",
			giveMethod: http.MethodGet,
			giveTarget: "/v1/users/1%27%20OR%20%271%27%3D%271%27%20--%20",
			wantCode:   http.StatusForbidden,
		},
		{
			name:       "xss in path",
			giveMethod: http.MethodGet,
			giveTarget: "/v1/users/%3Cscript%3Ealert(1)%3C%2Fscript%3E",
			wantCode:   http.StatusForbidden,
		},
		{
			name:       "sqli in query",
			giveMethod: http.MethodGet,
			giveTarget: "/v1/users?id=1%27%20OR%20%271%27%3D%271%27%20--%20",
			wantCode:   http.StatusForbidden,
		},
		{
			name:       "xss in query",
			giveMethod: http.MethodGet,
			giveTarget: "/v1/users?q=%3Cscript%3Ealert(1)%3C%2Fscript%3E",
			wantCode:   http.StatusForbidden,
		},
		{
			name:       "xss in user agent",
			giveMethod: http.MethodGet,
			giveTarget: "/v1/users",
			giveHeader: http.Header{"User-Agent": {_giveXSS}},
			wantCode:   http.StatusForbidden,
		},
		{
			name:       "sqli in cookie",
			giveMethod: http.MethodGet,
			giveTarget: "/v1/users",
			giveHeader: http.Header{
				"Cookie": {"sid=1%27 OR 1=1; theme=" + _giveSQLi},
			},
			wantCode: http.StatusForbidden,
		},
		{
			name:            "sqli in nested json body",
			giveMethod:      http.MethodPost,
			giveTarget:      "/v1/users",
			giveBody:        `{"user":{"names":["ok","` + _giveSQLi + `"]}}`,
			giveContentType: "application/json",
			wantCode:        http.StatusForbidden,
		},
		{
			name:            "xss in json body",
			giveMethod:      http.MethodPost,
			giveTarget:      "/v1/users",
			giveBody:        `{"name":"<script>alert(1)<\/script>"}`,
			giveContentType: "application/json; charset=utf-8",
			wantCode:        http.StatusForbidden,
		},
		{
			name:            "sqli in form body",
			giveMethod:      http.MethodPost,
			giveTarget:      "/v1/users",
			giveBody:        "name=1%27+OR+%271%27%3D%271%27+--+",
			giveContentType: "application/x-www-form-urlencoded",
			wantCode:        http.StatusForbidden,
		},
		{
			name:            "xss in raw body",
			giveMethod:      http.MethodPost,
			giveTarget:      "/v1/users",
			giveBody:        _giveXSS,
			giveContentType: "text/plain",
			wantCode:        http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				tt.giveMethod, tt.giveTarget, strings.NewReader(tt.giveBody))
			for key, vals := range tt.giveHeader {
				req.Header[key] = vals
			}
			req.Header.Set("Content-Type", tt.giveContentType)

			rec := httptest.NewRecorder()
			Injection(okHandler()).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
		})
	}
}

func TestInjection_ForbiddenPage(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(
		http.MethodPost, "/v1/users", strings.NewReader(_giveXSS))
	req.Header.Set("X-Request-Id", "req-123")

	rec := httptest.NewRecorder()
	Injection(okHandler()).ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8",
		rec.Header().Get("Content-Type"))
	assert.Equal(t, "req-123", rec.Header().Get("X-Request-Id"))
	assert.Contains(t, rec.Body.String(), "403 Forbidden")
	assert.Contains(t, rec.Body.String(), "req-123")
	assert.NotContains(t, rec.Body.String(), "handled")
}

func TestInjection_GeneratesRequestId(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(
		http.MethodPost, "/v1/users", strings.NewReader(_giveXSS))

	rec := httptest.NewRecorder()
	Injection(okHandler()).ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("X-Request-Id"))
}

func TestInjection_PreservesBody(t *testing.T) {
	t.Parallel()

	const give = `{"email":"a@b.co"}`

	var got string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		got = string(body)
	})

	req := httptest.NewRequest(
		http.MethodPost, "/v1/iam/login", strings.NewReader(give))
	req.Header.Set("Content-Type", "application/json")

	Injection(next).ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, give, got)
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("handled"))
	})
}
