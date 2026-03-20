// Copyright 2025 Sentinéz Labs.
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

package h

import (
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/setting/conf/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/h/logging"
	"github.com/sentinez/sentinez/internal/edge/v1/h/ratelimiter"
	"github.com/sentinez/sentinez/internal/edge/v1/h/routing"
	"github.com/sentinez/sentinez/internal/edge/v1/h/secure"
	"github.com/sentinez/sentinez/internal/edge/v1/h/static"
	"github.com/sentinez/sentinez/internal/edge/v1/h/trace"
	"github.com/sentinez/sentinez/internal/edge/v1/h/waitingroom"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
	"github.com/sentinez/shared/zlog"
)

func Init(appConf *confpb.Config) chains.Handler {
	var (
		hostname = appConf.GetEnv().GetHostname()
		ll       = zlog.LevelInfo
		curr     chains.Handler
		income   chains.Handler
	)
	// begin first middleware when request income
	income = trace.NewTracer()

	// current middleware
	curr = income

	curr = curr.SetNext(secure.NewDomain(hostname))

	curr = curr.SetNext(ratelimiter.New(ll))

	curr = curr.SetNext(waitingroom.New(ll))

	curr = curr.SetNext(static.NewStatic(ll))

	curr = curr.SetNext(logging.NewLogger(ll))

	curr = curr.SetNext(secure.NewRule(ll))

	curr = curr.SetNext(secure.NewWAF(ll))

	_ = curr.SetNext(routing.NewStandardRouter())

	return income
}
