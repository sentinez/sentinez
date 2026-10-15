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

package zlog

import (
	"context"
	"io"
	"testing"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

type recordExporter struct {
	records []sdklog.Record
}

func (e *recordExporter) Export(_ context.Context,
	rs []sdklog.Record) error {

	for _, r := range rs {
		e.records = append(e.records, r.Clone())
	}

	return nil
}

func (e *recordExporter) Shutdown(context.Context) error   { return nil }
func (e *recordExporter) ForceFlush(context.Context) error { return nil }

func TestLogCloserWithOTLPProvider(t *testing.T) {
	exp := &recordExporter{}
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(exp)),
	)

	l := NewLogCloser("test", typepb.LogType_LOG_TYPE_HTTP,
		LevelInfo, WithOTLPProvider(provider))
	l.Info("hello", &emptypb.Empty{}, io.NopCloser(nil))

	require.Len(t, exp.records, 1)
	require.Equal(t, "hello", exp.records[0].Body().AsString())
}

func TestLogCloserWithoutOTLP(t *testing.T) {
	exp := &recordExporter{}
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(exp)),
	)

	// WithoutOTLP wins over a provider given alongside it
	l := NewLogCloser("test", typepb.LogType_LOG_TYPE_HTTP,
		LevelInfo, WithOTLPProvider(provider), WithoutOTLP())
	l.Info("hello", &emptypb.Empty{}, io.NopCloser(nil))

	require.Empty(t, exp.records)
}

func TestOTLPConfigOf(t *testing.T) {
	tests := []struct {
		name         string
		giveEndpoint string
		wantEndpoint string
		wantInsecure bool
	}{
		{
			name:         "bare host",
			giveEndpoint: "localhost:4317",
			wantEndpoint: "localhost:4317",
			wantInsecure: true,
		},
		{
			name:         "http",
			giveEndpoint: "http://collector:4317",
			wantEndpoint: "collector:4317",
			wantInsecure: true,
		},
		{
			name:         "https",
			giveEndpoint: "https://collector:4317",
			wantEndpoint: "collector:4317",
			wantInsecure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := OTLPConfigOf(tt.giveEndpoint, "svc")

			assert.Equal(t, tt.wantEndpoint, cfg.Endpoint)
			assert.Equal(t, tt.wantInsecure, cfg.Insecure)
			assert.Equal(t, "svc", cfg.ServiceName)
		})
	}
}
