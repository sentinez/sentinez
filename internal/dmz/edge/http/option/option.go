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

// Package option holds the functional options shared by the edge chain
// handler and its chain nodes in internal/funcs.
package option

import (
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/shared/szlog"
	"go.opentelemetry.io/otel/sdk/log"
)

// Options is the resolved configuration of a chain node.
type Options struct {
	Named    string
	Level    szlog.Level
	Logger   *log.LoggerProvider
	Hostname string
}

// Option configures a chain node.
type Option interface {
	apply(*Options)
}

// New resolves opts on top of the defaults (edge service key as the log
// name, LevelInfo, no OTLP logger).
func New(opts ...Option) Options {
	o := Options{
		Named: edgepb.GetMetaEdgeServiceKey(),
		Level: szlog.LevelInfo,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}

	return o
}

// OTLP returns the szlog option exporting records through Logger; with a
// nil Logger the records are not exported.
func (o Options) OTLP() szlog.Option {
	if o.Logger == nil {
		return szlog.WithoutOTLP()
	}

	return szlog.WithOTLPProvider(o.Logger)
}

type namedOption string

func (n namedOption) apply(o *Options) {
	o.Named = string(n)
}

// WithNamed sets the logger name the node's records are written under.
func WithNamed(named string) Option {
	return namedOption(named)
}

type levelOption szlog.Level

func (l levelOption) apply(o *Options) {
	o.Level = szlog.Level(l)
}

// WithLevel sets the log level of the node.
func WithLevel(level szlog.Level) Option {
	return levelOption(level)
}

type loggerOption struct {
	provider *log.LoggerProvider
}

func (l loggerOption) apply(o *Options) {
	o.Logger = l.provider
}

// WithLogger exports the node's records through provider.
func WithLogger(provider *log.LoggerProvider) Option {
	return loggerOption{provider: provider}
}

type hostnameOption string

func (h hostnameOption) apply(o *Options) {
	o.Hostname = string(h)
}

// WithHostname sets the root domain tenants are served under.
func WithHostname(hostname string) Option {
	return hostnameOption(hostname)
}
