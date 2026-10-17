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

package ratelimiter

import (
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/shared/zlog"
)

func NewLimiter(
	store *memory.MemStore, opts ...option.Option) corechains.ChainNode {

	o := option.New(opts...)

	return &Limiter{
		Node:  corechains.NewNode(),
		store: store,
		log: zlog.NewLogCloser(
			o.Named,
			typepb.LogType_LOG_TYPE_RATE_LIMIT,
			o.Level,
			o.OTLP(),
		),
	}
}

type Limiter struct {
	*corechains.Node
	store *memory.MemStore
	log   zlog.LogCloser
}

// Handle counts the request against every limiter whose expression matches
// it, keyed by client IP. The request is rejected if any of them is over
// its limit.
func (l *Limiter) Handle(ctx corehttp.Context) error {
	key := string(ctx.RequestIP())

	for _, e := range l.store.Limiter().LoadContext(ctx) {
		if !e.Eval(ctx, nil) || e.Limiter.Allow(key) {
			continue
		}

		zlog.Debugf("edge: limiter %s exceeded %d/%d - %s",
			e.Rule.GetId(), e.Limiter.Count(key), e.Limiter.Limit(),
			ctx.URI())

		return corehttp.TooManyRequests(ctx)
	}

	return l.HandleNext(ctx)
}
