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

import "github.com/sentinez/core/szhttp"

// Handler handles a request.
type Handler interface {
	Handle(ctx szhttp.Context) error
}

// HandlerFunc adapts a function to a Handler.
type HandlerFunc func(ctx szhttp.Context) error

// Handle calls f(ctx).
func (f HandlerFunc) Handle(ctx szhttp.Context) error { return f(ctx) }

// Plugin is a plugin of a chain (Caddy-style): it is created as a zero
// value, pulls its dependencies in Provision, and receives the next
// handler on every call. It hands the request on by calling next, or
// answers it and returns without calling next.
type Plugin interface {
	// ID names the plugin, e.g. as the stage an error is tagged with.
	ID() string
	Handle(ctx szhttp.Context, next Handler) error
}

// Provisioner is implemented by a plugin that needs dependencies from an
// env of type E; it is called once by Build before the chain serves.
type Provisioner[E any] interface {
	Provision(env E) error
}

// Cleaner is implemented by a plugin holding resources to release when
// the chain is closed.
type Cleaner interface {
	Cleanup() error
}
