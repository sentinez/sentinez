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

package chains

import (
	"errors"
	"fmt"
	"slices"

	"github.com/sentinez/core/szhttp"
)

var (
	_ Handler = (*Chain)(nil)
	_ Handler = (*link)(nil)
)

// _end follows the last plugin; a request reaching it is left as the
// plugins answered it.
var _end = HandlerFunc(func(szhttp.Context) error { return nil })

// WrapFunc decorates the handler running the plugin id, e.g. to trace
// it.
type WrapFunc func(id string, h Handler) Handler

// Chain is a linked, provisioned list of plugins.
type Chain struct {
	head    Handler
	plugins []Plugin
}

// Build provisions plugins with env and links them in order, each one
// decorated by wrap when it is not nil. Only plugins implementing
// Provisioner[E] are provisioned. When a plugin fails to provision, the
// ones provisioned before it are cleaned up.
func Build[E any](env E, wrap WrapFunc, plugins ...Plugin) (*Chain, error) {
	plugins = slices.Clone(plugins)

	for i, pl := range plugins {
		p, ok := pl.(Provisioner[E])
		if !ok {
			continue
		}

		if err := p.Provision(env); err != nil {
			err = fmt.Errorf("provision %s: %w", pl.ID(), err)
			return nil, errors.Join(err, cleanup(plugins[:i]))
		}
	}

	var h Handler = _end
	for _, pl := range slices.Backward(plugins) {
		h = &link{plugin: pl, next: h}
		if wrap != nil {
			h = wrap(pl.ID(), h)
		}
	}

	return &Chain{head: h, plugins: plugins}, nil
}

// Handle runs the request through the chain.
func (c *Chain) Handle(ctx szhttp.Context) error {
	return c.head.Handle(ctx)
}

// Close cleans the plugins up, in reverse order.
func (c *Chain) Close() error {
	return cleanup(c.plugins)
}

// link runs plugin with the handler following it.
type link struct {
	plugin Plugin
	next   Handler
}

func (l *link) Handle(ctx szhttp.Context) error {
	return l.plugin.Handle(ctx, l.next)
}

func cleanup(plugins []Plugin) error {
	var err error
	for _, pl := range slices.Backward(plugins) {
		if c, ok := pl.(Cleaner); ok {
			err = errors.Join(err, c.Cleanup())
		}
	}

	return err
}
