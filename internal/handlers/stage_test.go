// Copyright 2026 Sentinéz Labs.
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
	"net/http"
	"testing"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// panicker panics with its value instead of handling the request.
type panicker struct{ value func() }

func (*panicker) ID() string { return "PAN" }

func (p *panicker) Handle(szhttp.Context, chains.Handler) error {
	p.value()
	return nil
}

func TestStageRecoversPanic(t *testing.T) {
	var empty []int
	tests := []struct {
		name string
		give func() // panics
	}{
		{name: "string", give: func() { panic("boom") }},
		{name: "error", give: func() { panic(errFake) }},
		{name: "runtime error", give: func() { _ = empty[1] }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log []string
			chain, err := chains.Build(newEnv(), Stage,
				&fake{id: "A", log: &log},
				&panicker{value: tt.give},
			)
			require.NoError(t, err)

			ctx := newContext()
			log = nil
			require.NoError(t, chain.Handle(ctx))

			assert.Equal(t, http.StatusInternalServerError, ctx.StatusCode())
			assert.Equal(t, []string{"handle A"}, log)
		})
	}
}

func TestStageRepanicsAbort(t *testing.T) {
	chain, err := chains.Build(newEnv(), Stage, &panicker{
		value: func() { panic(http.ErrAbortHandler) },
	})
	require.NoError(t, err)

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		_ = chain.Handle(newContext())
	})
}
