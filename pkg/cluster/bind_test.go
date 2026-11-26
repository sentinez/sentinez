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
	"errors"
	"maps"
	"testing"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// _clustered the env of an edge in a cluster.
var _clustered = map[string]string{
	"SZ_MEMBERSHIP_ADDRESS": "10.0.0.1:3322",
	"SZ_DISCOVERY_ADDRESS":  "0.0.0.0:3322",
}

// privateIPOf returns ip as the private IP of the host.
func privateIPOf(ip string) privateIPFunc {
	return func() (string, error) { return ip, nil }
}

// clusteredWith the env of an edge in a cluster, with env.
func clusteredWith(env map[string]string) map[string]string {
	all := maps.Clone(_clustered)
	maps.Copy(all, env)

	return all
}

type olricBindCase struct {
	name     string
	give     map[string]string
	wantHost string
	wantPort int
}

var _olricBindCases = []olricBindCase{
	{name: "alone", wantHost: "127.0.0.1", wantPort: 3320},
	{name: "clustered", give: _clustered, wantHost: "10.0.0.2",
		wantPort: 3320},
	{
		name:     "set",
		give:     map[string]string{"SZ_OLRIC_ADDRESS": "10.0.0.5:3330"},
		wantHost: "10.0.0.5",
		wantPort: 3330,
	},
	{
		name:     "port only",
		give:     map[string]string{"SZ_OLRIC_ADDRESS": ":3330"},
		wantHost: "10.0.0.2",
		wantPort: 3330,
	},
	{
		name: "all interfaces",
		give: clusteredWith(map[string]string{
			"SZ_OLRIC_ADDRESS": "0.0.0.0:3320",
		}),
		wantHost: "10.0.0.2",
		wantPort: 3320,
	},
	{
		name:     "public ip set",
		give:     map[string]string{"SZ_OLRIC_ADDRESS": "203.0.113.7:3320"},
		wantHost: "203.0.113.7",
		wantPort: 3320,
	},
}

func TestOlricBind(t *testing.T) {
	for _, tt := range _olricBindCases {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := olricBind(&settingpb.Config{Env: tt.give},
				privateIPOf("10.0.0.2"))
			require.NoError(t, err)

			assert.Equal(t, tt.wantHost, host)
			assert.Equal(t, tt.wantPort, port)
		})
	}
}

func TestOlricBindError(t *testing.T) {
	tests := []struct {
		name string
		give map[string]string
	}{
		{name: "no port", give: map[string]string{"SZ_OLRIC_ADDRESS": "e1"}},
		{
			name: "bad port",
			give: map[string]string{"SZ_OLRIC_ADDRESS": "e1:70000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := olricBind(&settingpb.Config{Env: tt.give},
				privateIPOf("10.0.0.2"))
			require.Error(t, err)
		})
	}
}

func TestOlricBindWithoutPrivateIP(t *testing.T) {
	tests := []struct {
		name string
		give map[string]string
	}{
		{name: "clustered", give: _clustered},
		{
			name: "all interfaces",
			give: map[string]string{"SZ_OLRIC_ADDRESS": "[::]:3320"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := olricBind(&settingpb.Config{Env: tt.give},
				privateIPOf(""))
			require.ErrorIs(t, err, errNoPrivateIP)
		})
	}
}

func TestOlricBindPrivateIPError(t *testing.T) {
	failing := func() (string, error) { return "", errors.New("no ifaces") }

	_, _, err := olricBind(&settingpb.Config{Env: _clustered}, failing)

	require.Error(t, err)
}

func TestDiscoveryBind(t *testing.T) {
	tests := []struct {
		name     string
		give     string
		wantHost string
		wantPort int
	}{
		{name: "all interfaces", give: "0.0.0.0:7777", wantHost: "10.0.0.2",
			wantPort: 7777},
		{name: "set", give: "10.0.0.9:7777", wantHost: "10.0.0.9",
			wantPort: 7777},
		{name: "invalid", give: "nope", wantHost: "10.0.0.2",
			wantPort: 3322},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &settingpb.Config{Env: map[string]string{
				"SZ_DISCOVERY_ADDRESS": tt.give,
			}}

			host, port, err := discoveryBind(conf, privateIPOf("10.0.0.2"))
			require.NoError(t, err)

			assert.Equal(t, tt.wantHost, host)
			assert.Equal(t, tt.wantPort, port)
		})
	}
}

func TestDiscoveryBindWithoutPrivateIP(t *testing.T) {
	conf := &settingpb.Config{Env: map[string]string{
		"SZ_DISCOVERY_ADDRESS": "nope",
	}}

	_, _, err := discoveryBind(conf, privateIPOf(""))

	require.ErrorIs(t, err, errNoPrivateIP)
}
