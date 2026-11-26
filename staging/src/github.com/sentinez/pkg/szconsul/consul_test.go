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

package szconsul

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		give string
	}{
		{name: "default agent", give: ""},
		{name: "http", give: "http://consul:8500"},
		{name: "https", give: "https://consul:8501"},
		{name: "unix socket", give: "unix:///run/consul.sock"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.give)
			require.NoError(t, err)
		})
	}
}

func TestNewBadURL(t *testing.T) {
	tests := []struct {
		name string
		give string
	}{
		{name: "bare address", give: "consul:8500"},
		{name: "other scheme", give: "tcp://consul:8500"},
		{name: "two hosts", give: "http://a:8500,b:8500"},
		{name: "no host", give: "http://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.give)
			require.Error(t, err)
		})
	}
}
