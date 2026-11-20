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
	"net/http"
	"strconv"
	"testing"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/sentinez/internal/handlers/wrap/stage"
	"github.com/stretchr/testify/require"
)

// pass hands the request on.
type pass struct{ id string }

func (p *pass) ID() string { return p.id }

func (*pass) Handle(ctx szhttp.Context, next chains.Handler) error {
	return next.Handle(ctx)
}

// deny answers the request with a 403, as a rule blocking it.
type deny struct{}

func (*deny) ID() string { return "DNY" }

func (*deny) Handle(ctx szhttp.Context, _ chains.Handler) error {
	ctx.SetStatusCode(http.StatusForbidden)
	return nil
}

// BenchmarkStageDenied a request answered 403 by the last of the stages
// of a chain as long as the edge one: each stage sees the error status.
func BenchmarkStageDenied(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	n := len(plugins())
	chained := make([]chains.Plugin, 0, n)
	for i := range n - 1 {
		chained = append(chained, &pass{id: strconv.Itoa(i)})
	}

	chain, err := chains.Build(newEnv(), stage.Wrap,
		append(chained, &deny{})...)
	require.NoError(b, err)

	ctx := newContext()
	b.ReportAllocs()

	for b.Loop() {
		_ = chain.Handle(ctx)
	}
}
