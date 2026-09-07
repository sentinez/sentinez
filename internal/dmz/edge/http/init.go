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
	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/setting/v1"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/cdn"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/logging"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/ratelimiter"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/room"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/routing"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/secure"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/static"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/trace"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/shared/zlog"
)

func Init(appConf *settingpb.Config,
	memStore *memory.MemStore) corechains.ChainNode {
	var (
		hostname = appConf.Get(settingpb.Senz_SENZ_HOSTNAME)
		ll       = zlog.LevelInfo
		curr     corechains.ChainNode
		income   corechains.ChainNode
	)
	// begin first middleware when request income
	income = trace.NewTracer(ll)

	// current middleware
	curr = income

	curr = curr.SetNext(logging.NewLogger(ll, memStore))

	curr = curr.SetNext(secure.NewDomainBased(hostname, memStore))

	curr = curr.SetNext(cdn.NewCache(ll, memStore))

	curr = curr.SetNext(ratelimiter.NewLimiter(ll, memStore))

	curr = curr.SetNext(room.NewRoom(ll, memStore))

	curr = curr.SetNext(static.NewStatic(ll, memStore))

	curr = curr.SetNext(secure.NewRuleBased(ll, memStore))

	curr = curr.SetNext(secure.NewWAF(ll, memStore))

	_ = curr.SetNext(routing.NewStandardRouter(memStore))

	return income
}
