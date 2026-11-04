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

package server

import (
	"context"

	"github.com/sentinez/core/szrunner"
	"github.com/sentinez/sentinez/pkg/cluster/leader"
)

// Leader campaigns for the leadership of the edge zone and, while it
// leads, keeps the zone Status in Consul (leader.PublishStatus).
func Leader(c *szrunner.Context[leader.Elector]) {
	c.Inject(leader.New)

	c.Serve(func(ctx context.Context, e *leader.Elector) error {
		return e.Run(ctx, e.PublishStatus())
	})
}
