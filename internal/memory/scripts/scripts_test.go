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

package scripts

import (
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	s := New()
	good := &edgepb.Script{
		Id:     "a",
		Source: "---@stage before/router\nfunction f(c) end",
	}
	off := &edgepb.Script{Id: "b", Source: "syntax error (", Disabled: true}

	require.NoError(t, s.Store("ns", []*edgepb.Script{good, off}))
	assert.Len(t, s.Load("ns"), 1, "a disabled script is not compiled")

	first := s.Load("ns")[0]
	require.NoError(t, s.Store("ns", []*edgepb.Script{good}))
	assert.Same(t, first, s.Load("ns")[0], "an unchanged script is kept")

	bad := &edgepb.Script{Id: "c", Source: "function ("}
	require.Error(t, s.Store("ns", []*edgepb.Script{bad}))
	assert.Len(t, s.Load("ns"), 1, "the previous scripts stay")

	require.NoError(t, s.Store("ns", nil))
	assert.Empty(t, s.Load("ns"))

	assert.Empty(t, (*Scripts)(nil).LoadContext(nil))
}
