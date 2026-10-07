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

package memory

import (
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/memory/ratelimiter"
)

func TestLoadRateLimiterKeepsAllLimiters(t *testing.T) {
	m := &MemStore{limiter: ratelimiter.New()}
	err := m.LoadRateLimiter(&edgepb.Setting{
		Server: &edgepb.Server{Name: "multi-limiter"},
		Security: &edgepb.Security{Limiters: []*edgepb.RateLimit{
			newRateLimit("a", typepb.Status_STATUS_ACTIVE, 1),
			newRateLimit("b", typepb.Status_STATUS_DISABLE, 1),
			newRateLimit("c", typepb.Status_STATUS_ACTIVE, 2),
		}},
	})
	if err != nil {
		t.Fatalf("LoadRateLimiter: %v", err)
	}

	entries := m.Limiter().Load("multi-limiter")

	var got []string
	for _, e := range entries {
		got = append(got, e.Rule.GetId())
	}

	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Fatalf("limiters = %v, want [a c]", got)
	}

	if entries[0].Limiter == entries[1].Limiter {
		t.Error("limiters share the same RateLimiter")
	}
}

func newRateLimit(
	id string, status typepb.Status, limit int64,
) *edgepb.RateLimit {
	return &edgepb.RateLimit{
		IngressRuntime: &secrulepb.SecRule{Id: id, Status: status},
		TimeWindow:     "1s",
		Timeout:        "1s",
		Limit:          limit,
	}
}
