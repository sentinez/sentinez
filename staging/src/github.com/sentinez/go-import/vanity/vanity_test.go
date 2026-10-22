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

package vanity

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerMeta(t *testing.T) {
	tests := []struct {
		name       string
		givePath   string
		wantImport string
	}{
		{
			name:     "module root",
			givePath: "/core?go-get=1",
			wantImport: `content="go.sentinez.com/core git ` +
				`https://github.com/sentinez/core"`,
		},
		{
			name:     "nested package",
			givePath: "/shared/szlog?go-get=1",
			wantImport: `content="go.sentinez.com/shared git ` +
				`https://github.com/sentinez/shared"`,
		},
		{
			name:     "submodule in subdirectory",
			givePath: "/contrib/httphz?go-get=1",
			wantImport: `content="go.sentinez.com/contrib git ` +
				`https://github.com/sentinez/contrib"`,
		},
	}

	h := New(DefaultConfig())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := serve(t, h, tt.givePath)
			if !strings.Contains(body, tt.wantImport) {
				t.Errorf("missing go-import %s in:\n%s", tt.wantImport, body)
			}
		})
	}
}

func TestHandlerInfo(t *testing.T) {
	tests := []struct {
		name     string
		givePath string
		wantText string
	}{
		{
			name:     "index lists repos",
			givePath: "/",
			wantText: `<a href="/controlplane">controlplane</a>`,
		},
		{
			name:     "repo page",
			givePath: "/shared/szlog",
			wantText: "go get go.sentinez.com/shared/szlog",
		},
		{
			name:     "go-get other than 1",
			givePath: "/core?go-get=0",
			wantText: "https://pkg.go.dev/go.sentinez.com/core",
		},
	}

	h := New(DefaultConfig())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := serve(t, h, tt.givePath)
			if strings.Contains(body, `name="go-import"`) {
				t.Errorf("info page has go-import meta:\n%s", body)
			}
			if !strings.Contains(body, tt.wantText) {
				t.Errorf("missing %s in:\n%s", tt.wantText, body)
			}
		})
	}
}

func TestHandlerStatus(t *testing.T) {
	tests := []struct {
		name       string
		giveMethod string
		givePath   string
		wantCode   int
	}{
		{"logo", http.MethodGet, "/_static/logo.png", 200},
		{"icon", http.MethodGet, "/_static/icon.png", 200},
		{"missing static", http.MethodGet, "/_static/nope.png", 404},
		{"unknown repo meta", http.MethodGet, "/nope?go-get=1", 404},
		{"unknown repo info", http.MethodGet, "/nope", 404},
		{"post rejected", http.MethodPost, "/core", 405},
	}

	h := New(DefaultConfig())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.giveMethod, tt.givePath, nil)
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func serve(t *testing.T, h *Handler, path string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	return rec.Body.String()
}
