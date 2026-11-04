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

package cluster

import (
	"testing"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigReportsStarted(t *testing.T) {
	membership := settingpb.Senz_SENZ_MEMBERSHIP_ADDRESS.String()
	discovery := settingpb.Senz_SENZ_DISCOVERY_ADDRESS.String()

	tests := []struct {
		name string
		give map[string]string
	}{
		{name: "standalone"},
		{
			name: "clustered",
			give: map[string]string{
				membership: "10.0.0.1:3322",
				discovery:  "0.0.0.0:3322",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var started bool
			conf := newConfig(&settingpb.Config{Env: tt.give},
				func() { started = true })

			require.NotNil(t, conf.Started)
			conf.Started()
			assert.True(t, started)
		})
	}
}

func TestMarkReadyIsIdempotent(t *testing.T) {
	c := &Cluster{ready: make(chan struct{})}

	c.markReady()
	c.markReady()

	select {
	case <-c.Ready():
	default:
		t.Fatal("Ready not closed")
	}
}
