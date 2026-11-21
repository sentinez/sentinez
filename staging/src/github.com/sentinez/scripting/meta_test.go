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

package scripting

import (
	"os"
	"regexp"
	"slices"
	"testing"

	"github.com/sentinez/scripting/internal/api"
	"github.com/sentinez/scripting/internal/hooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetaMatchesContext the annotations of meta/sentinez.lua declare the
// methods ctx has, the functions of the log table and the stages, no
// more, no less.
func TestMetaMatchesContext(t *testing.T) {
	src, err := os.ReadFile("meta/sentinez.lua")
	require.NoError(t, err)

	methods := declared(src, `(?m)^function Context:(\w+)\(`)
	assert.Equal(t, api.CtxMethods(), methods)

	logs := declared(src, `(?m)^function log\.(\w+)\(`)
	assert.Equal(t, api.LogFuncs(), logs)

	stages := declared(src, `(?m)^---\| "(\w+)"`)
	assert.Equal(t, slices.Sorted(slices.Values(hooks.StageNames())), stages)
}

// declared the sorted names the pattern captures in src.
func declared(src []byte, pattern string) []string {
	var names []string
	for _, m := range regexp.MustCompile(pattern).FindAllSubmatch(src, -1) {
		names = append(names, string(m[1]))
	}

	return slices.Sorted(slices.Values(names))
}
