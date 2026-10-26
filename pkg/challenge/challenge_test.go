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

package challenge

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		giveKey string
		wantErr bool
	}{
		{"base64 key", "c2VjcmV0", false},
		{"empty key", "", false},
		{"invalid base64", "not base64!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.giveKey)

			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestIsVerify(t *testing.T) {
	tests := []struct {
		name       string
		giveMethod string
		givePath   string
		want       bool
	}{
		{"post challenge path", http.MethodPost, ChallengePath, true},
		{"get challenge path", http.MethodGet, ChallengePath, false},
		{"post other path", http.MethodPost, "/login", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.giveMethod, tt.givePath, nil)
			ctx := httpx.NewContext(req, httptest.NewRecorder())

			assert.Equal(t, tt.want, IsVerify(ctx))
		})
	}
}

func TestLocalPath(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"/shop?page=2", "/shop?page=2"},
		{"/", "/"},
		{"", "/"},
		{"https://evil.example", "/"},
		{"//evil.example", "/"},
		{`/\evil.example`, "/"},
		{"/a\r\nSet-Cookie: x=1", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, LocalPath(tt.give))
		})
	}
}
