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
	giveKeys    []settingpb.Sz
	wantFound   map[string]string
	wantSkipped []string
}

type kv = map[string]string

var _overlayCases = []overlayCase{
	{
		name:      "a shared key",
		giveKv:    kv{"SZ_HOSTNAME": "s6z.io.vn"},
		wantFound: kv{"SZ_HOSTNAME": "s6z.io.vn"},
	},
	{
		name: "the folder of the service wins over the shared key",
		giveKv: kv{
			"SZ_KAFKA_URI":          "kafka://shared:9094",
			_edge + "/SZ_KAFKA_URI": "kafka://edge:9094",
		},
		wantFound: kv{"SZ_KAFKA_URI": "kafka://edge:9094"},
	},
	{
		name:      "the folder of another service is ignored",
		giveKv:    kv{"sentinez.gateway.apiserver.v1/SZ_HOSTNAME": "x"},
		wantFound: kv{},
	},
	{
		name:    "the environment wins over both",
		giveEnv: kv{"SZ_KAFKA_URI": "kafka://env:9094"},
		giveKv: kv{
			"SZ_KAFKA_URI":          "kafka://shared:9094",
			_edge + "/SZ_KAFKA_URI": "kafka://edge:9094",
		},
		wantFound: kv{},
	},
	{
		name:      "an empty environment value does not win",
		giveEnv:   kv{"SZ_HOSTNAME": ""},
		giveKv:    kv{"SZ_HOSTNAME": "s6z.io.vn"},
		wantFound: kv{"SZ_HOSTNAME": "s6z.io.vn"},
	},
	{
		name:      "an empty value is ignored",
		giveKv:    kv{"SZ_HOSTNAME": ""},
		wantFound: kv{},
	},
	{
		name: "keys that only the environment may set are refused",
		giveKv: kv{
			"SZ_CONSUL_URI":         "x",
			"SZ_ADDRESS":            "x",
			"SZ_MEMBERSHIP_ADDRESS": "x",
			"SZ_DISCOVERY_ADDRESS":  "x",
			"SZ_OLRIC_ADDRESS":      "x",
			_edge + "/SZ_ZONE":      "x",
		},
		wantFound: kv{},
		wantSkipped: []string{
			"SZ_ADDRESS", "SZ_CONSUL_URI", "SZ_DISCOVERY_ADDRESS",
			"SZ_MEMBERSHIP_ADDRESS", "SZ_OLRIC_ADDRESS", _edge + "/SZ_ZONE",
		},
	},
	{
		name:        "an unknown key is refused",
		giveKv:      kv{"SZ_NOPE": "x", "SZ_UNSPECIFIED": "x"},
		wantFound:   kv{},
		wantSkipped: []string{"SZ_NOPE", "SZ_UNSPECIFIED"},
	},
	{
		name:      "a key the service does not read is ignored",
		giveKv:    kv{"SZ_HOSTNAME": "x", "SZ_KAFKA_URI": "k"},
		giveKeys:  []settingpb.Sz{settingpb.Sz_SZ_KAFKA_URI},
		wantFound: kv{"SZ_KAFKA_URI": "k"},
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
