// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package corecmn

import (
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
)

func TestNormalizeEdgeSettingLimiters(t *testing.T) {
	tests := []struct {
		name        string
		giveIngress *secrulepb.SecRuleLite
		wantStatus  typepb.Status
	}{
		{
			name:       "no ingress applies to all",
			wantStatus: typepb.Status_STATUS_ACTIVE,
		},
		{
			name:        "active ingress",
			giveIngress: &secrulepb.SecRuleLite{Status: "STATUS_ACTIVE"},
			wantStatus:  typepb.Status_STATUS_ACTIVE,
		},
		{
			name:        "disabled ingress",
			giveIngress: &secrulepb.SecRuleLite{Status: "STATUS_DISABLE"},
			wantStatus:  typepb.Status_STATUS_DISABLE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeLimiterRule(tt.giveIngress)
			if got == nil {
				t.Fatal("IngressRuntime is nil")
			}

			if got.GetStatus() != tt.wantStatus {
				t.Errorf("status = %v, want %v", got.GetStatus(), tt.wantStatus)
			}
		})
	}
}

func normalizeLimiterRule(ingress *secrulepb.SecRuleLite) *secrulepb.SecRule {
	lim := &edgepb.RateLimit{Ingress: ingress}
	NormalizeEdgeSetting(&edgepb.Setting{
		Security: &edgepb.Security{
			Limiters: []*edgepb.RateLimit{lim},
		},
	})

	return lim.GetIngressRuntime()
}

func TestNormalizeEdgeSettingLimiterDurations(t *testing.T) {
	// Durations are [timeWindow, timeout].
	tests := []struct {
		name string
		give [2]string
		want [2]string
	}{
		{
			name: "missing both uses default",
			want: [2]string{"5s", "5s"},
		},
		{
			name: "missing timeout only",
			give: [2]string{"1m", ""},
			want: [2]string{"1m", "5s"},
		},
		{
			name: "set values are kept",
			give: [2]string{"1m", "30s"},
			want: [2]string{"1m", "30s"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lim := &edgepb.RateLimit{
				TimeWindow: tt.give[0],
				Timeout:    tt.give[1],
			}
			normalizeLimiter(lim)

			got := [2]string{lim.GetTimeWindow(), lim.GetTimeout()}
			if got != tt.want {
				t.Errorf("durations = %q, want %q", got, tt.want)
			}
		})
	}
}
