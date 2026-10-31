// Copyright 2025 Duc-Hung Ho.
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

package trace

import (
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/shared/szprefix"
	"github.com/sentinez/shared/szrand"
)

// TracerId is the chain stage name of Tracer.
const TracerId = "TRC"

var _ chains.Plugin = (*Tracer)(nil)

// Tracer makes sure the request has an id.
type Tracer struct{}

func (*Tracer) ID() string { return TracerId }

// Handle makes sure the request has an id. The engine context assigns it
// once when the request is accepted; one is only generated here for a
// context that has none, so a request never gets a second id.
func (*Tracer) Handle(ctx szhttp.Context, next chains.Handler) error {
	if ctx.RequestId() == "" {
		ctx.SetRequestId(szrand.NewULID(szprefix.Request))
	}

	return next.Handle(ctx)
}
