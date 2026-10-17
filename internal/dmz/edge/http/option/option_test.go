// Copyright 2026 Sentinéz Labs.
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

package option

import (
	"testing"

	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/shared/zlog"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/log"
)

func TestNewDefaults(t *testing.T) {
	want := Options{
		Named: edgepb.GetMetaEdgeServiceKey(),
		Level: zlog.LevelInfo,
	}

	assert.Equal(t, want, New())
}

func TestNew(t *testing.T) {
	provider := log.NewLoggerProvider()
	tests := []struct {
		name     string
		giveOpts []Option
		want     Options
	}{
		{
			name: "all options",
			giveOpts: []Option{
				WithNamed("edge-test"),
				WithLevel(zlog.LevelError),
				WithLogger(provider),
				WithHostname("is.s6z.io.vn"),
			},
			want: Options{
				Named:    "edge-test",
				Level:    zlog.LevelError,
				Logger:   provider,
				Hostname: "is.s6z.io.vn",
			},
		},
		{
			name: "last option wins",
			giveOpts: []Option{
				WithNamed("a"), WithNamed("b"), WithLevel(zlog.LevelError),
			},
			want: Options{Named: "b", Level: zlog.LevelError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.giveOpts...))
		})
	}
}
