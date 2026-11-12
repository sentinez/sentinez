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

func TestEnvKeysAlwaysHoldTheConsulUri(t *testing.T) {
	p := New(Spec{Keys: []settingpb.Senz{settingpb.Senz_SENZ_HOSTNAME}})

	assert.Equal(t, []settingpb.Senz{
		settingpb.Senz_SENZ_HOSTNAME, settingpb.Senz_SENZ_CONSUL_URI,
	}, p.envKeys())
	assert.Equal(t, []settingpb.Senz{settingpb.Senz_SENZ_HOSTNAME}, p.spec.Keys)
}

func TestEnvKeysWithTheConsulUri(t *testing.T) {
	keys := []settingpb.Senz{settingpb.Senz_SENZ_CONSUL_URI}

	assert.Equal(t, keys, New(Spec{Keys: keys}).envKeys())
}

func TestEnvKeysOfAllKeys(t *testing.T) {
	assert.Nil(t, New(Spec{}).envKeys())
}
