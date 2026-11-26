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
	membership := settingpb.Sz_SZ_MEMBERSHIP_ADDRESS.String()
	discovery := settingpb.Sz_SZ_DISCOVERY_ADDRESS.String()

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
			conf, err := newConfig(&settingpb.Config{Env: tt.give},
				func() { started = true }, privateIPOf("10.0.0.2"))
			require.NoError(t, err)

			require.NotNil(t, conf.Started)
			conf.Started()
			assert.True(t, started)
		})
	}
}

func TestNewConfigBindsPrivateIPs(t *testing.T) {
	conf, err := newConfig(&settingpb.Config{Env: _clustered},
		func() {}, privateIPOf("10.0.0.2"))
	require.NoError(t, err)

	assert.Equal(t, "10.0.0.2", conf.BindAddr)
	assert.Equal(t, 3320, conf.BindPort)
	assert.Equal(t, "10.0.0.2", conf.MemberlistConfig.BindAddr)
	assert.Equal(t, 3322, conf.MemberlistConfig.BindPort)
	assert.Equal(t, []string{"10.0.0.1:3322"}, conf.Peers)
}

func TestNewConfigWithoutPrivateIP(t *testing.T) {
	_, err := newConfig(&settingpb.Config{Env: _clustered},
		func() {}, privateIPOf(""))

	require.ErrorIs(t, err, errNoPrivateIP)
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
