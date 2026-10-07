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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otellog "go.opentelemetry.io/otel/log"
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
// call) forwards its records to it, unless it was given its own provider with
// WithOTLPProvider. The returned function flushes pending records and shuts
// the exporter down; call it on service exit.
func SetupOTLP(ctx context.Context,
	cfg OTLPConfig) (func(context.Context) error, error) {
	provider, err := NewOTLPProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}
	otel.SetLoggerProvider(provider)

	return provider.Shutdown, nil
}

// NewOTLPProvider builds an OTLP/gRPC LoggerProvider without installing it
// globally. Hand it to a single logger with WithOTLPProvider; the caller owns
// the provider and must Shutdown it to flush pending records.
func NewOTLPProvider(ctx context.Context,
	cfg OTLPConfig) (*sdklog.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(ctx, exporterOptions(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("zlog: create otlp log exporter: %w", err)
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	), nil
}

func exporterOptions(cfg OTLPConfig) []otlploggrpc.Option {
	var opts []otlploggrpc.Option
	if cfg.Endpoint != "" {
		opts = append(opts, otlploggrpc.WithEndpoint(cfg.Endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(cfg.Headers))
	}

	return opts
}

// newResource merges the configured service attributes into the default
// resource.
func newResource(ctx context.Context,
	cfg OTLPConfig) (*resource.Resource, error) {
	var attrs []attribute.KeyValue
	if cfg.ServiceName != "" {
		attrs = append(attrs, semconv.ServiceName(cfg.ServiceName))
	}
	if cfg.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.ServiceVersion))
	}
	if len(attrs) == 0 {
		return resource.Default(), nil
	}

	extra, err := resource.New(ctx, resource.WithAttributes(attrs...))
	if err != nil {
		return nil, fmt.Errorf("zlog: create otlp resource: %w", err)
	}

	res, err := resource.Merge(resource.Default(), extra)
	if err != nil {
		return nil, fmt.Errorf("zlog: merge otlp resource: %w", err)
	}

	return res, nil
}

// Option customizes a logger created by NewLog or NewLogCloser.
type Option func(*loggingOptions)

type loggingOptions struct {
	provider otellog.LoggerProvider
}

// WithOTLPProvider exports this logger's records through p instead of the
// global LoggerProvider, so it can use its own OTLP endpoint and resource.
func WithOTLPProvider(p otellog.LoggerProvider) Option {
	return func(o *loggingOptions) { o.provider = p }
}

func newLoggingOptions(opts []Option) loggingOptions {
	var o loggingOptions
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// withOTLP tees the core with an OpenTelemetry bridge. Without a provider the
// bridge resolves the global LoggerProvider lazily, so it is a no-op until
// SetupOTLP is called.
func withOTLP(core zapcore.Core, scope string,
	p otellog.LoggerProvider) zapcore.Core {

	if p == nil {
		return zapcore.NewTee(core, otelzap.NewCore(scope))
	}

	return zapcore.NewTee(core,
		otelzap.NewCore(scope, otelzap.WithLoggerProvider(p)))
}
