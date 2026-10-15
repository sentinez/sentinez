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

package http

import (
	corechains "github.com/sentinez/core/http/chains"
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/sentinez/sentinez/internal/funcs/cdn"
	"github.com/sentinez/sentinez/internal/funcs/logging"
	"github.com/sentinez/sentinez/internal/funcs/ratelimiter"
	"github.com/sentinez/sentinez/internal/funcs/room"
	"github.com/sentinez/sentinez/internal/funcs/routing"
	"github.com/sentinez/sentinez/internal/funcs/secure"
	"github.com/sentinez/sentinez/internal/funcs/static"
	"github.com/sentinez/sentinez/internal/funcs/trace"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/shared/zlog"
	"go.opentelemetry.io/otel/sdk/log"
)

// nolint
func NewHandler(appConf *settingpb.Config, ms *memory.MemStore, logx *log.LoggerProvider) corechains.ChainNode {
	var (
		hostname = appConf.Get(settingpb.Senz_SENZ_HOSTNAME)
		level    = zlog.LevelInfo
		curr     corechains.ChainNode
		income   corechains.ChainNode
	)

	// begin first middleware when request income
	income = trace.NewTracer(level)

	// current middleware
	curr = income

	curr = curr.SetNext(trace.Wrap("LOG", logging.NewLogging(level, logx)))

	curr = curr.SetNext(trace.Wrap("DMA", secure.NewDomainBased(hostname, ms)))

	curr = curr.SetNext(trace.Wrap("PUB", cdn.NewPublic(level, ms)))

	curr = curr.SetNext(trace.Wrap("CDN", cdn.NewCache(level, ms)))

	curr = curr.SetNext(trace.Wrap("LMT", ratelimiter.NewLimiter(level, ms)))

	curr = curr.SetNext(trace.Wrap("ROM", room.NewRoom(level, ms)))

	curr = curr.SetNext(trace.Wrap("STC", static.NewStatic(level, ms)))

	curr = curr.SetNext(trace.Wrap("RUL", secure.NewSecRule(level, ms)))

	curr = curr.SetNext(trace.Wrap("WAF", secure.NewWAF(level, ms)))

	_ = curr.SetNext(trace.Wrap("ROU", routing.NewStandardRouter(ms)))

	return income
}
