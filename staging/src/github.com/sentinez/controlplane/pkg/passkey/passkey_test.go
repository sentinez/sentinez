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

package passkey

import (
	"testing"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
)

func TestClientOrigins(t *testing.T) {
	tests := []struct {
		name string
		give string
		want []string
	}{
		{"unset", "", nil},
		{"one", "https://a.io", []string{"https://a.io"}},
		{"several", "https://a.io, https://b.io:3000",
			[]string{"https://a.io", "https://b.io:3000"}},
		{"trailing slash and empty items", "https://a.io/,,",
			[]string{"https://a.io"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &settingpb.Config{Env: map[string]string{
				settingpb.Senz_SENZ_CLIENT_ORIGIN.String(): tt.give,
			}}

			assert.Equal(t, tt.want, ClientOrigins(conf))
		})
	}
}
