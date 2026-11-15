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

package main

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/sentinez/controlplane/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Setenv("SENZ_SERVICE_KEY",
		base64.StdEncoding.EncodeToString([]byte("service-secret")))
	var out bytes.Buffer

	err := run([]string{"-name", "edge-1", "-ttl", "1h"}, &out)

	require.NoError(t, err)
	got, ok := crypto.ServiceTokenVerifier(strings.TrimSpace(out.String()))
	require.True(t, ok)
	assert.Equal(t, "edge-1", got.GetName())
}

func TestRunRejects(t *testing.T) {
	tests := []struct {
		name     string
		giveArgs []string
	}{
		{"no name", []string{}},
		{"bad ttl", []string{"-name", "e", "-ttl", "-1h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.giveArgs, &bytes.Buffer{})

			assert.Error(t, err)
		})
	}
}

func TestRunWithoutKey(t *testing.T) {
	t.Setenv("SENZ_SERVICE_KEY", "")

	err := run([]string{"-name", "edge-1"}, &bytes.Buffer{})

	assert.Error(t, err)
}
