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
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSetting(name string) *edgepb.Setting {
	return &edgepb.Setting{Server: &edgepb.Server{Name: name}}
}

func TestVersion(t *testing.T) {
	va, err := Version(newSetting("a"))
	require.NoError(t, err)
	again, err := Version(newSetting("a"))
	require.NoError(t, err)
	vb, err := Version(newSetting("b"))
	require.NoError(t, err)

	assert.Len(t, va, _versionLen)
	assert.Equal(t, va, again)
	assert.NotEqual(t, va, vb)
}
