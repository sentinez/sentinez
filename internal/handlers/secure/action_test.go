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

package secure

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestApplyActionsHeaders(t *testing.T) {
	ctx := newTestContext(http.MethodGet, "/", "", nil)

	applyActions(ctx, []*rulepb.Action{
		newAction(t, _modifyHeader, map[string]any{
			"X-Tier": "gold", "X-Score": 7, "Bad Key": "x",
		}),
		newAction(t, _modifyHeader, map[string]any{"X-Tier": "silver"}),
	})

	assert.Equal(t, "gold", string(ctx.Header([]byte("X-Tier"))),
		"the higher priority rule wins")
	assert.Equal(t, "7", string(ctx.Header([]byte("X-Score"))))
	assert.Empty(t, ctx.Header([]byte("Bad Key")))
}

func TestApplyActionsTags(t *testing.T) {
	ctx := newTestContext(http.MethodGet, "/", "", nil)

	applyActions(ctx, []*rulepb.Action{
		newAction(t, _tag, map[string]any{"risk": "high", "bot": true}),
		newAction(t, _tag, map[string]any{"risk": "low", "team": "a"}),
	})

	want := map[string]string{"risk": "high", "bot": "true", "team": "a"}
	assert.Equal(t, want, ctx.X().GetTags())
}

func TestApplyActionsRoute(t *testing.T) {
	ctx := newTestContext(http.MethodGet, "/", "", nil)

	applyActions(ctx, []*rulepb.Action{
		newAction(t, _routeTo, map[string]any{"location": "/api"}),
		newAction(t, _routeTo, map[string]any{"location": "/local"}),
		newAction(t, _routeTo, nil),
	})

	assert.Equal(t, "/api", ctx.X().GetRouteLocation())
}

func TestRedirect(t *testing.T) {
	tests := []struct {
		name         string
		giveParams   map[string]any
		wantStatus   int
		wantLocation string
	}{
		{"default status", map[string]any{"location": "/new"},
			http.StatusFound, "/new"},
		{"permanent", map[string]any{"location": "https://x.example/",
			"status": 301}, http.StatusMovedPermanently, "https://x.example/"},
		{"status as string", map[string]any{"location": "/new",
			"status": "307"}, http.StatusTemporaryRedirect, "/new"},
		{"not a redirect status", map[string]any{"location": "/new",
			"status": 200}, http.StatusFound, "/new"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx := newTestContext(http.MethodGet, "/", "", rec)

			done, err := redirect(ctx, newAction(t, _redirect, tt.giveParams))

			require.NoError(t, err)
			assert.True(t, done)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantLocation, rec.Header().Get("Location"))
		})
	}
}

func TestRedirectUnusable(t *testing.T) {
	tests := []struct {
		name       string
		giveParams map[string]any
	}{
		{"no location", nil},
		{"header injection", map[string]any{"location": "/a\r\nX: y"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ctx := newTestContext(http.MethodGet, "/", "", rec)

			done, err := redirect(ctx, newAction(t, _redirect, tt.giveParams))

			require.NoError(t, err)
			assert.False(t, done)
			assert.Empty(t, rec.Header().Get("Location"))
		})
	}
}

func TestLoggedAction(t *testing.T) {
	block := &rulepb.Action{Type: _block}
	log := &rulepb.Action{Type: _log}
	tag := &rulepb.Action{Type: _tag}

	assert.Equal(t, block, loggedAction(block, []*rulepb.Action{log}))
	assert.Equal(t, log, loggedAction(nil, []*rulepb.Action{log, tag}))
	assert.Nil(t, loggedAction(nil, nil))
}

func newAction(
	t *testing.T, typ rulepb.ActionType, params map[string]any,
) *rulepb.Action {
	t.Helper()

	s, err := structpb.NewStruct(params)
	require.NoError(t, err)

	return &rulepb.Action{Type: typ, Params: s}
}
