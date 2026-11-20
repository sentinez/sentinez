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

package secure

import (
	"net/http"
	"testing"
	"time"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/core/szlimiter"
	"github.com/sentinez/core/szrule"
	"github.com/sentinez/pkg/szlog"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/sentinez/sentinez/internal/memory/secratelimits"
)

// BenchmarkSecLimiterAllowed a request counted by a limiter of its
// namespace, under the limit.
func BenchmarkSecLimiterAllowed(b *testing.B) {
	szlog.SetLogLevel(szlog.LevelInfo)

	const namespace = "bench.example.com"

	expr := &rulepb.Expression{OrCondition: []*rulepb.AndCondition{{}}}
	store := secratelimits.New()
	store.Store(namespace, []secratelimits.Entry{{
		Eval:    szrule.NewEval("srlim_bench", expr),
		Rule:    &secrulepb.SecRule{Id: "srlim_bench", Expr: expr},
		Limiter: szlimiter.NewRateLimiter(time.Second, time.Second, 1<<62),
	}})

	ctx := newTestContext(http.MethodGet, "/", "", nil)
	ctx.X().Namespace = namespace
	next := chains.HandlerFunc(func(szhttp.Context) error { return nil })
	plugin := &SecLimiter{store: store}
	b.ReportAllocs()

	for b.Loop() {
		_ = plugin.Handle(ctx, next)
	}
}
