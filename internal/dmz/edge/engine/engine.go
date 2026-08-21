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

package engine

import (
	"context"

	"github.com/sentinez/contrib/httphz"
	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/core/runner"
	"github.com/sentinez/sentinez/internal/dmz/edge/transport"
	"github.com/sentinez/sentinez/pkg/apps/dmz/edge"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
)

func Hertz(c *runner.Context[edge.Server]) {
	c.Inject(httphz.NewServer)

	c.OnStart(func(_ context.Context, server *edge.Server) error {
		server.SetOptions(corehttp.WithOnConnect(transport.OnHertzConnect))

		return nil
	})
}

func Standard(c *runner.Context[edge.Server]) {
	c.Inject(stdhttpx.NewServer)

	c.OnStart(func(_ context.Context, server *edge.Server) error {
		server.SetOptions(corehttp.WithOnConnect(transport.OnStandardConnect))

		return nil
	})
}
