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

package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		giveURL string
		wantErr bool
	}{
		{name: "default agent", giveURL: ""},
		{name: "http", giveURL: "http://127.0.0.1:8500"},
		{name: "bad scheme", giveURL: "ftp://127.0.0.1:8500", wantErr: true},
		{name: "no host", giveURL: "http://", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dcvr, err := New(WithConsulURL(tt.giveURL))

			assert.Equal(t, tt.wantErr, err != nil, err)
			assert.Equal(t, tt.wantErr, dcvr == nil)
		})
	}
}

func TestNewErrorWraps(t *testing.T) {
	_, err := New(WithConsulURL("ftp://127.0.0.1:8500"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "discovery: consul uri")
}
