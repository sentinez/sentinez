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

package envplugin

import (
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/shared/szlog"
)

// Env holds what plugins provision themselves from: the options are
// resolved once for the whole chain.
type Env struct {
	opts option.Options
	mem  *memory.MemStore
}

// NewEnv returns an Env serving mem; opts are resolved with option.New.
func NewEnv(mem *memory.MemStore, opts ...option.Option) *Env {
	return &Env{opts: option.New(opts...), mem: mem}
}

// Store returns the memory store the plugins read rules from.
func (e *Env) Store() *memory.MemStore {
	return e.mem
}

// Hostname returns the root domain tenants are served under.
func (e *Env) Hostname() string {
	return e.opts.Hostname
}

// Logger returns an event logger of kind, named and leveled by the
// options and exported through their OTLP provider.
func (e *Env) Logger(kind typepb.LogType) szlog.LogCloser {
	return szlog.NewLogCloser(
		e.opts.Named, kind, e.opts.Level, e.opts.OTLP())
}
