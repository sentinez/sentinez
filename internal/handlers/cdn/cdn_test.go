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

package cdn

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/sentinez/sentinez/internal/memory/cdnrules"
	"github.com/sentinez/sentinez/pkg/challenge"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheSkipsSystemPaths runs a Cache without a rule store: reaching
// the store or the cache would panic, so passing means both were skipped.
func TestCacheSkipsSystemPaths(t *testing.T) {
	tests := []struct {
		name     string
		givePath string
	}{
		{"public image", PublicPrefix + "sentinez/1/1.png"},
		{"challenge", challenge.ChallengePath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.givePath, nil)
			ctx := httpx.NewContext(req, httptest.NewRecorder())

			called := false
			next := chains.HandlerFunc(func(szhttp.Context) error {
				called = true
				return nil
			})

			require.NoError(t, (&Cache{}).Handle(ctx, next))
			assert.True(t, called)
		})
	}
}

func TestMatchAny(t *testing.T) {
	yes := cdnrules.Entry{
		Eval: func(szhttp.Request, *rulepb.MatchedRules) bool { return true },
	}
	no := cdnrules.Entry{
		Eval: func(szhttp.Request, *rulepb.MatchedRules) bool { return false },
	}

	tests := []struct {
		name      string
		giveChain []cdnrules.Entry
		want      bool
	}{
		{"no rules", nil, false},
		{"none match", []cdnrules.Entry{no, no}, false},
		{"a later rule matches", []cdnrules.Entry{no, yes}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := httpx.NewContext(req, httptest.NewRecorder())

			assert.Equal(t, tt.want, matchAny(ctx, tt.giveChain))
		})
	}
}
