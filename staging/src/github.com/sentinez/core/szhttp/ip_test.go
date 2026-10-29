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

func TestStripPort(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{name: "ipv4", give: "127.0.0.1", want: "127.0.0.1"},
		{name: "ipv4 port", give: "127.0.0.1:44835", want: "127.0.0.1"},
		{name: "ipv6", give: "2001:db8::1", want: "2001:db8::1"},
		{name: "ipv6 loopback", give: "::1", want: "::1"},
		{name: "ipv6 bracket", give: "[2001:db8::1]", want: "2001:db8::1"},
		{name: "ipv6 port", give: "[2001:db8::1]:443", want: "2001:db8::1"},
		{name: "ipv4 mapped", give: "::ffff:8.8.8.8", want: "::ffff:8.8.8.8"},
		{name: "unclosed", give: "[2001:db8::1", want: "[2001:db8::1"},
		{name: "empty", give: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(StripPort([]byte(tt.give))))
		})
	}
}

func TestStripPortAllocs(t *testing.T) {
	addr := []byte("[2001:db8::1]:443")

	allocs := testing.AllocsPerRun(100, func() { _ = StripPort(addr) })

	assert.Zero(t, allocs)
}
