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

package szsdklog

import (
	"context"
	"testing"

	settingpb "github.com/sentinez/sentinez/api/proto/sentinez/types/setting/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	kafka := settingpb.Senz_SENZ_KAFKA_BROKERS.String()
	otlp := settingpb.Senz_SENZ_OTLP_EXTERNAL_ENDPOINT.String()
	tests := []struct {
		name      string
		giveEnv   map[string]string
		wantNil   bool
		wantKafka bool
		wantOTLP  bool
	}{
		{name: "none", wantNil: true},
		{
			name:     "otlp",
			giveEnv:  map[string]string{otlp: "localhost:4317"},
			wantOTLP: true,
		},
		{
			name:      "kafka wins over otlp",
			giveEnv:   map[string]string{kafka: "b:9092", otlp: "c:4317"},
			wantKafka: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := New(&settingpb.Config{Env: tt.giveEnv})
			require.NoError(t, err)
			t.Cleanup(func() { _ = e.Shutdown(context.Background()) })

			assert.Equal(t, tt.wantNil, e == nil)
			assert.Equal(t, tt.wantKafka, e != nil && e.kafka != nil)
			assert.Equal(t, tt.wantOTLP, e != nil && e.otlp != nil)
			assert.NotNil(t, e.Option())
		})
	}
}

func TestNilExporter(t *testing.T) {
	var e *Exporter

	assert.NotNil(t, e.Option())
	assert.NoError(t, e.Shutdown(context.Background()))
}
