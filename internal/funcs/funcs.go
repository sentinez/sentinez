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

package funcs

import (
	corechains "github.com/sentinez/core/http/chains"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/funcs/cdn"
	"github.com/sentinez/sentinez/internal/funcs/logging"
	"github.com/sentinez/sentinez/internal/funcs/ratelimiter"
	"github.com/sentinez/sentinez/internal/funcs/room"
	"github.com/sentinez/sentinez/internal/funcs/routing"
	"github.com/sentinez/sentinez/internal/funcs/secure"
	"github.com/sentinez/sentinez/internal/funcs/static"
	"github.com/sentinez/sentinez/internal/funcs/trace"
	"github.com/sentinez/sentinez/internal/memory"
)

// New returns a Function whose nodes read mem and are configured by
// opts (see option.WithLevel, option.WithLogger, option.WithHostname).
func New(mem *memory.MemStore, opts ...option.Option) []corechains.ChainNode {
	return []corechains.ChainNode{
		trace.NewTracer(opts...),
		trace.Wrap("LOG", logging.NewLogging(opts...)),
		trace.Wrap("DMA", secure.NewDomainBased(mem, opts...)),
		trace.Wrap("PUB", cdn.NewPublic(mem, opts...)),
		trace.Wrap("CDN", cdn.NewCache(mem, opts...)),
		trace.Wrap("LMT", ratelimiter.NewLimiter(mem, opts...)),
		trace.Wrap("ROM", room.NewRoom(mem, opts...)),
		trace.Wrap("STC", static.NewStatic(mem, opts...)),
		trace.Wrap("RUL", secure.NewSecRule(mem, opts...)),
		trace.Wrap("WAF", secure.NewWAF(mem, opts...)),
		trace.Wrap("ROU", routing.NewStandardRouter(mem)),
	}
}
