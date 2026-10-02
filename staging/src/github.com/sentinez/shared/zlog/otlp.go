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
	"fmt"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap/zapcore"
)

// OTLPConfig configures the OpenTelemetry Protocol (OTLP/gRPC) log exporter.
type OTLPConfig struct {
	// Endpoint is the collector address, e.g. "localhost:4317".
	// When empty, the standard OTEL_EXPORTER_OTLP_* env vars are honored.
	Endpoint string
	// Insecure disables transport security.
	Insecure bool
	// Headers are sent with every export request.
	Headers map[string]string
	// ServiceName is reported as the service.name resource attribute.
	ServiceName string
	// ServiceVersion is reported as the service.version resource attribute.
	ServiceVersion string
}

// SetupOTLP installs a global OpenTelemetry LoggerProvider that exports logs
// over OTLP/gRPC. Every logger created by this package (before or after this
// call) forwards its records to it. The returned function flushes pending
// records and shuts the exporter down; call it on service exit.
func SetupOTLP(ctx context.Context, cfg OTLPConfig) (func(context.Context) error, error) {
	opts := []otlploggrpc.Option{}
	if cfg.Endpoint != "" {
		opts = append(opts, otlploggrpc.WithEndpoint(cfg.Endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(cfg.Headers))
	}

	exporter, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("zlog: create otlp log exporter: %w", err)
	}

	attrs := resource.Default()
	if cfg.ServiceName != "" || cfg.ServiceVersion != "" {
		var extra []resource.Option
		if cfg.ServiceName != "" {
			extra = append(extra, resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)))
		}
		if cfg.ServiceVersion != "" {
			extra = append(extra, resource.WithAttributes(semconv.ServiceVersion(cfg.ServiceVersion)))
		}
		res, err := resource.New(ctx, extra...)
		if err != nil {
			return nil, fmt.Errorf("zlog: create otlp resource: %w", err)
		}
		if attrs, err = resource.Merge(attrs, res); err != nil {
			return nil, fmt.Errorf("zlog: merge otlp resource: %w", err)
		}
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(attrs),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
	otel.SetLoggerProvider(provider)

	return provider.Shutdown, nil
}

// withOTLP tees the core with an OpenTelemetry bridge. The bridge resolves the
// global LoggerProvider lazily, so it is a no-op until SetupOTLP is called.
func withOTLP(core zapcore.Core, scope string) zapcore.Core {
	return zapcore.NewTee(core, otelzap.NewCore(scope))
}
