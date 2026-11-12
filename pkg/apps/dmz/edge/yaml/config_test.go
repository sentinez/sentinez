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

package edgeyaml

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sentinez/sentinez/api/client/consul"
	"github.com/sentinez/sentinez/internal/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeGetter struct {
	data []byte
	err  error
	key  string
}

func (f *fakeGetter) GetRaw(_ context.Context, key string) ([]byte, error) {
	f.key = key

	return f.data, f.err
}

func fileWith(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "proxy.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func TestReadProxyFromConsul(t *testing.T) {
	getter := &fakeGetter{data: []byte("from consul")}

	data, err := readProxy(getter, "/no/such/file")

	require.NoError(t, err)
	assert.Equal(t, "from consul", string(data))
	assert.Equal(t, defaults.ProxyConfigKey, getter.key)
}

func TestReadProxyFileWhenConsulHasNoKey(t *testing.T) {
	getter := &fakeGetter{err: consul.ErrKeyNotFound}

	data, err := readProxy(getter, fileWith(t, "from file"))

	require.NoError(t, err)
	assert.Equal(t, "from file", string(data))
}

func TestReadProxyConsulFails(t *testing.T) {
	// a Consul that fails is not a reason to serve an old file
	getter := &fakeGetter{err: errors.New("connection refused")}

	_, err := readProxy(getter, fileWith(t, "from file"))

	assert.ErrorContains(t, err, "connection refused")
}
