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

package appconfig

import (
	"testing"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
)

const _edge = "sentinez.dmz.edge.v1"

type overlayCase struct {
	name        string
	giveEnv     map[string]string
	giveKv      map[string]string
	giveKeys    []settingpb.Senz
	wantFound   map[string]string
	wantSkipped []string
}

type kv = map[string]string

var _overlayCases = []overlayCase{
	{
		name:      "a shared key",
		giveKv:    kv{"SENZ_HOSTNAME": "s6z.io.vn"},
		wantFound: kv{"SENZ_HOSTNAME": "s6z.io.vn"},
	},
	{
		name: "the folder of the service wins over the shared key",
		giveKv: kv{
			"SENZ_KAFKA_BROKERS":          "shared:9094",
			_edge + "/SENZ_KAFKA_BROKERS": "edge:9094",
		},
		wantFound: kv{"SENZ_KAFKA_BROKERS": "edge:9094"},
	},
	{
		name:      "the folder of another service is ignored",
		giveKv:    kv{"sentinez.gateway.apiserver.v1/SENZ_HOSTNAME": "x"},
		wantFound: kv{},
	},
	{
		name:    "the environment wins over both",
		giveEnv: kv{"SENZ_KAFKA_BROKERS": "env:9094"},
		giveKv: kv{
			"SENZ_KAFKA_BROKERS":          "shared:9094",
			_edge + "/SENZ_KAFKA_BROKERS": "edge:9094",
		},
		wantFound: kv{},
	},
	{
		name:      "an empty environment value does not win",
		giveEnv:   kv{"SENZ_HOSTNAME": ""},
		giveKv:    kv{"SENZ_HOSTNAME": "s6z.io.vn"},
		wantFound: kv{"SENZ_HOSTNAME": "s6z.io.vn"},
	},
	{
		name:      "an empty value is ignored",
		giveKv:    kv{"SENZ_HOSTNAME": ""},
		wantFound: kv{},
	},
	{
		name: "keys that only the environment may set are refused",
		giveKv: kv{
			"SENZ_CONSUL_URI":         "x",
			"SENZ_ADDRESS":            "x",
			"SENZ_MEMBERSHIP_ADDRESS": "x",
			"SENZ_DISCOVERY_ADDRESS":  "x",
			_edge + "/SENZ_ZONE":      "x",
		},
		wantFound: kv{},
		wantSkipped: []string{
			"SENZ_ADDRESS", "SENZ_CONSUL_URI", "SENZ_DISCOVERY_ADDRESS",
			"SENZ_MEMBERSHIP_ADDRESS", _edge + "/SENZ_ZONE",
		},
	},
	{
		name:        "an unknown key is refused",
		giveKv:      kv{"SENZ_NOPE": "x", "SENZ_UNSPECIFIED": "x"},
		wantFound:   kv{},
		wantSkipped: []string{"SENZ_NOPE", "SENZ_UNSPECIFIED"},
	},
	{
		name:      "a key the service does not read is ignored",
		giveKv:    kv{"SENZ_HOSTNAME": "x", "SENZ_KAFKA_BROKERS": "k"},
		giveKeys:  []settingpb.Senz{settingpb.Senz_SENZ_KAFKA_BROKERS},
		wantFound: kv{"SENZ_KAFKA_BROKERS": "k"},
	},
}

func TestOverlay(t *testing.T) {
	for _, tt := range _overlayCases {
		t.Run(tt.name, func(t *testing.T) {
			found, skipped := overlay(tt.giveEnv, tt.giveKv, _edge, tt.giveKeys)

			assert.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.wantSkipped, skipped)
		})
	}
}
