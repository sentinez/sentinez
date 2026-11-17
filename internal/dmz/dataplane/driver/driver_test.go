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

package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The inputs are checked before the maps are touched: a Context without
// loaded objects refuses them, no kernel needed.

func TestLookupBandwidthRefusesNonIPv4(t *testing.T) {
	tests := []struct {
		name   string
		giveIP string
	}{
		{name: "ipv6", giveIP: "2001:db8::1"},
		{name: "not an ip", giveIP: "nope"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (&Context{}).LookupBandwidth(tt.giveIP)

			assert.ErrorIs(t, err, errNotIPv4)
		})
	}
}

func TestBlockCIDRRefusesNonIPv4(t *testing.T) {
	err := (&Context{}).BlockCIDR("2001:db8::/32")

	assert.ErrorIs(t, err, errNotIPv4)
}

func TestBlockCIDRRefusesMalformed(t *testing.T) {
	err := (&Context{}).BlockCIDR("10.0.0.0/33")

	assert.ErrorContains(t, err, "driver: invalid CIDR")
}
