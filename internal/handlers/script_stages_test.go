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

package handlers

import (
	"slices"
	"testing"

	"github.com/sentinez/scripting/szlua"
	"github.com/sentinez/sentinez/internal/handlers/secure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScriptStagesAreTheChain the stages a script may hook are the
// plugins of the chain after DMA, which resolves the namespace the
// scripts are looked up by: a plugin added, removed or renamed there has
// to be in szlua.Stages too, or no hook reaches it.
func TestScriptStagesAreTheChain(t *testing.T) {
	ids := make([]string, 0, len(plugins()))
	for _, p := range plugins() {
		ids = append(ids, p.ID())
	}

	dma := slices.Index(ids, secure.DomainBasedId)
	require.GreaterOrEqual(t, dma, 0)

	assert.Equal(t, ids[dma+1:], szlua.Stages())
}
