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

package consul

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Name    string `json:"name"    toml:"name"`
	Enabled bool   `json:"enabled" toml:"enabled"`
}

// fakeKV serves the subset of the Consul KV HTTP API the client uses.
type fakeKV struct {
	mu   sync.Mutex
	data map[string][]byte
}

type kvPair struct {
	Key   string `json:"Key"`
	Value []byte `json:"Value"`
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	srv := httptest.NewServer(&fakeKV{data: make(map[string][]byte)})
	t.Cleanup(srv.Close)

	c, err := New(strings.TrimPrefix(srv.URL, "http://"))
	require.NoError(t, err)
	return c
}

func (f *fakeKV) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := strings.TrimPrefix(r.URL.Path, "/v1/kv/")
	switch r.Method {
	case http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		f.data[key] = body
		_, _ = w.Write([]byte("true"))
	case http.MethodDelete:
		delete(f.data, key)
		_, _ = w.Write([]byte("true"))
	default:
		f.serveGet(w, key, r.URL.Query().Has("recurse"))
	}
}

func (f *fakeKV) serveGet(w http.ResponseWriter, key string, recurse bool) {
	var pairs []kvPair
	for k, v := range f.data {
		if k == key || (recurse && strings.HasPrefix(k, key)) {
			pairs = append(pairs, kvPair{Key: k, Value: v})
		}
	}
	if len(pairs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key < pairs[j].Key
	})
	_ = json.NewEncoder(w).Encode(pairs)
}

func TestGetString(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Put(ctx, "app/name", "sentinez"))

	got, err := Get[string](ctx, c, "app/name")
	require.NoError(t, err)
	assert.Equal(t, "sentinez", got)
}

func TestGetBytes(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Put(ctx, "app/raw", []byte{0x1, 0x2}))

	got, err := Get[[]byte](ctx, c, "app/raw")
	require.NoError(t, err)
	assert.Equal(t, []byte{0x1, 0x2}, got)
}

func TestGetStruct(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	want := testConfig{Name: "edge", Enabled: true}
	require.NoError(t, c.Put(ctx, "app/conf", want))

	got, err := Get[testConfig](ctx, c, "app/conf")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetInt(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Put(ctx, "app/limit", 42))

	got, err := Get[int](ctx, c, "app/limit")
	require.NoError(t, err)
	assert.Equal(t, 42, got)
}

func TestGetNotFound(t *testing.T) {
	c := newTestClient(t)

	_, err := Get[string](context.Background(), c, "missing")
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestGetDecodeError(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Put(ctx, "app/limit", "not-a-number"))

	_, err := Get[int](ctx, c, "app/limit")
	assert.Error(t, err)
}

func TestGetOr(t *testing.T) {
	c := newTestClient(t)

	got, err := GetOr(context.Background(), c, "missing", 10)
	require.NoError(t, err)
	assert.Equal(t, 10, got)
}

func TestDelete(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Put(ctx, "app/name", "sentinez"))
	require.NoError(t, c.Delete(ctx, "app/name"))

	_, err := c.GetRaw(ctx, "app/name")
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestList(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Put(ctx, "tenants/", nil))
	require.NoError(t, c.Put(ctx, "tenants/a", testConfig{Name: "a"}))
	require.NoError(t, c.Put(ctx, "tenants/b", testConfig{Name: "b"}))

	got, err := List[testConfig](ctx, c, "tenants/")
	require.NoError(t, err)
	assert.Equal(t, map[string]testConfig{
		"a": {Name: "a"},
		"b": {Name: "b"},
	}, got)
}

func TestResolveFormat(t *testing.T) {
	tests := []struct {
		name     string
		giveKey  string
		giveOpts []Option
		want     Format
	}{
		{name: "default", giveKey: "app/conf", want: FormatJSON},
		{name: "json ext", giveKey: "app/conf.json", want: FormatJSON},
		{name: "toml ext", giveKey: "app/conf.toml", want: FormatTOML},
		{name: "upper ext", giveKey: "app/conf.TOML", want: FormatTOML},
		{
			name:     "option wins",
			giveKey:  "app/conf.toml",
			giveOpts: []Option{WithFormat(FormatJSON)},
			want:     FormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveFormat(tt.giveKey, tt.giveOpts)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPutTOMLByExtension(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	want := testConfig{Name: "edge", Enabled: true}
	require.NoError(t, c.Put(ctx, "app/conf.toml", want))

	raw, err := Get[string](ctx, c, "app/conf.toml")
	require.NoError(t, err)
	assert.Equal(t, "name = 'edge'\nenabled = true\n", raw)

	got, err := Get[testConfig](ctx, c, "app/conf.toml")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetTOMLWithFormat(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	raw := "name = \"edge\"\nenabled = true\n"
	require.NoError(t, c.Put(ctx, "app/conf", raw))

	got, err := Get[testConfig](ctx, c, "app/conf", WithFormat(FormatTOML))
	require.NoError(t, err)
	assert.Equal(t, testConfig{Name: "edge", Enabled: true}, got)
}

func TestGetTOMLDecodeError(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Put(ctx, "app/conf.toml", "name = "))

	_, err := Get[testConfig](ctx, c, "app/conf.toml")
	assert.Error(t, err)
}

func TestListMixedFormats(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	require.NoError(t, c.Put(ctx, "tenants/a.toml", testConfig{Name: "a"}))
	require.NoError(t, c.Put(ctx, "tenants/b.json", testConfig{Name: "b"}))

	got, err := List[testConfig](ctx, c, "tenants/")
	require.NoError(t, err)
	assert.Equal(t, map[string]testConfig{
		"a.toml": {Name: "a"},
		"b.json": {Name: "b"},
	}, got)
}
