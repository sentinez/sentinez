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

package szhttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseContextIds(t *testing.T) {
	tests := []struct {
		name         string
		giveTenant   string
		giveResource string
		wantTenant   string
		wantResource string
	}{
		{"unset", "", "", "tnt_d", "rsrc_d"},
		{"empty keeps default", "", "r1", "tnt_d", "r1"},
		{"set", "t1", "r1", "t1", "r1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBaseContext("tnt_d", "rsrc_d")

			b.SetTenantId(tt.giveTenant)
			b.SetResourceId(tt.giveResource)

			assert.Equal(t, tt.wantTenant, b.TenantId())
			assert.Equal(t, tt.wantResource, b.ResourceId())
		})
	}
}

func TestBaseContextResetKeepsDefaults(t *testing.T) {
	b := NewBaseContext("tnt_d", "rsrc_d")
	b.SetTenantId("t1")

	b.Reset()

	assert.Equal(t, "tnt_d", b.TenantId())
	assert.Equal(t, "rsrc_d", b.ResourceId())
}
