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

const _scopeQuery = "?q=%3Cscript%3Ealert(1)%3C%2Fscript%3E"

var _scopeTests = []struct {
	name       string
	giveTarget string
	wantCode   int
}{
	{
		name:       "scoped module root",
		giveTarget: "/iam" + _scopeQuery,
		wantCode:   http.StatusForbidden,
	},
	{
		name:       "scoped module route",
		giveTarget: "/iam/users" + _scopeQuery,
		wantCode:   http.StatusForbidden,
	},
	{
		name:       "second scoped module",
		giveTarget: "/tenant/resources" + _scopeQuery,
		wantCode:   http.StatusForbidden,
	},
	{
		name:       "clean request in scope",
		giveTarget: "/iam/users?q=admin",
		wantCode:   http.StatusOK,
	},
	{
		name:       "module out of scope",
		giveTarget: "/security/secrules" + _scopeQuery,
		wantCode:   http.StatusOK,
	},
	{
		name:       "prefix is not a full segment",
		giveTarget: "/iamx/users" + _scopeQuery,
		wantCode:   http.StatusOK,
	},
}

func TestScope(t *testing.T) {
	t.Parallel()

	for _, tt := range _scopeTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.giveTarget, nil)
			rec := httptest.NewRecorder()

			Scope(Injection, "/iam", "/tenant")(okHandler()).
				ServeHTTP(rec, req)

			assert.Equal(t, tt.wantCode, rec.Code)
		})
	}
}
