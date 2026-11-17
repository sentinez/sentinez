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

package headers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func forwarded(t *testing.T, value string) context.Context {
	t.Helper()

	return metadata.NewIncomingContext(t.Context(),
		metadata.Pairs(_forwardedFor, value))
}

func TestClientAddr(t *testing.T) {
	tests := []struct {
		name string
		give context.Context
		want string
	}{
		{"one hop", forwarded(t, "10.0.0.1"), "10.0.0.1"},
		{"the last of several", forwarded(t, "6.6.6.6, 10.0.0.1"),
			"10.0.0.1"},
		{"nothing known", t.Context(), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ClientAddr(tt.give))
		})
	}
}

func TestUserAgent(t *testing.T) {
	tests := []struct {
		name string
		give metadata.MD
		want string
	}{
		{"the one of the HTTP request first", metadata.Pairs(
			_gatewayUserAgent, "curl/8", _userAgent, "grpc-go"), "curl/8"},
		{"else the one of the gRPC client",
			metadata.Pairs(_userAgent, "grpc-go"), "grpc-go"},
		{"none", metadata.MD{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(t.Context(), tt.give)
			assert.Equal(t, tt.want, UserAgent(ctx))
		})
	}
}
