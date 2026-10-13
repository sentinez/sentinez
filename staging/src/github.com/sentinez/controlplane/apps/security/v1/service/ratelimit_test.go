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

package securitysvc

import (
	"testing"

	securitypb "github.com/sentinez/sentinez/api/proto/sentinez/apps/security/v1"
	"github.com/stretchr/testify/assert"
)

type rl = securitypb.RateLimit

func TestValidateRateLimit_Valid(t *testing.T) {
	tests := []struct {
		name string
		give *rl
	}{
		{"no timeout", &rl{TimeWindow: "10s", MaxRequests: 50}},
		{"timeout", &rl{TimeWindow: "1m", MaxRequests: 1, Timeout: "30s"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, validateRateLimit(tt.give))
		})
	}
}

func TestValidateRateLimit_Invalid(t *testing.T) {
	tests := []struct {
		name string
		give *rl
	}{
		{"nil", nil},
		{"zero max requests", &rl{TimeWindow: "10s"}},
		{"missing time window", &rl{MaxRequests: 10}},
		{"negative window", &rl{TimeWindow: "-5s", MaxRequests: 10}},
		{"zero timeout", &rl{
			TimeWindow: "5s", MaxRequests: 10, Timeout: "0s",
		}},
		{"invalid timeout", &rl{
			TimeWindow: "5s", MaxRequests: 10, Timeout: "abc",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, validateRateLimit(tt.give))
		})
	}
}
