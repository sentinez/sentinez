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

// Package exchange pools the Exchange the edge hands the Lua hooks of a
// stage, per request hooked.
package exchange

import (
	"io"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/variable"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/pkg/szsync"
	scriptex "github.com/sentinez/scripting/exchange"
)

var (
	_ io.Closer         = (*Exchange)(nil)
	_ scriptex.Exchange = (*Exchange)(nil)

	_pool = szsync.NewPool[Exchange]()
)

// Exchange the request of a hook, released on Close: the headers it sets
// are validated, and the ones sent upstream are recorded for the egress
// log.
type Exchange struct {
	szhttp.Context
}

// Acquire returns an Exchange of the pool on ctx.
func Acquire(ctx szhttp.Context) *Exchange {
	e := _pool.Get()
	e.Context = ctx

	return e
}

// Release resets obj and puts it back in the pool.
func Release(obj *Exchange) {
	obj.Context = nil
	_pool.Put(obj)
}

// Close releases e to the pool.
func (e *Exchange) Close() error {
	Release(e)

	return nil
}

// SetHeader sets a header of the request sent upstream.
func (e *Exchange) SetHeader(k, v string) {
	if variable.IsValidHeaderKey(k) {
		szhttp.SetUpstreamHeader(e.Context, k, v)
	}
}

// SetResponseHeader sets a header of the response.
func (e *Exchange) SetResponseHeader(k, v string) {
	if variable.IsValidHeaderKey(k) {
		e.Context.SetResponseHeader([]byte(k), []byte(v))
	}
}

// Log writes a line of log.debug or log.info of a script, with the id of
// the request.
func (e *Exchange) Log(level scriptex.Level, script, msg string) {
	rid := e.RequestId()
	if level == scriptex.LevelDebug {
		szlog.Debugf("script: request %s: %s: %s", rid, script, msg)

		return
	}

	szlog.Infof("script: request %s: %s: %s", rid, script, msg)
}
