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

package funcs

import (
	"errors"
	"net/http"
	"time"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/core/sztracer"
	"github.com/sentinez/shared/szlog"
)

var (
	_ chains.WrapFunc = Stage
	_ chains.Handler  = (*stage)(nil)
)

// Stage decorates h, the handler running the plugin name, so that an
// error returned from it (and not already attributed to a deeper stage)
// is tagged with name and logged once. Because errors unwind from the
// innermost plugin outwards, the first stage to see an error is the one
// that produced it.
func Stage(name string, h chains.Handler) chains.Handler {
	return &stage{name: name, node: h}
}

type stage struct {
	name string
	node chains.Handler
}

func (s *stage) Handle(ctx szhttp.Context) error {
	begin := time.Now()
	err := s.node.Handle(ctx)
	if err == nil {
		s.traceStatus(ctx, begin)
		return nil
	}

	var se *sztracer.StageError
	if errors.As(err, &se) {
		return err
	}

	se = &sztracer.StageError{
		Stage: s.name, Elapsed: time.Since(begin), Err: err,
	}
	szlog.Errorf("tracer: request %s failed at stage %q status %d after %s: %v",
		ctx.RequestId(), se.Stage, ctx.StatusCode(), se.Elapsed, err)

	return se
}

// traceStatus logs a request answered with an error status without an
// error, e.g. a plugin that rendered a 403 page. Errors unwind from the
// innermost plugin outwards, so for a request the first stage logged is
// the one that answered; the stages after it only passed it through.
func (s *stage) traceStatus(ctx szhttp.Context, begin time.Time) {
	status := ctx.StatusCode()
	if status < http.StatusBadRequest {
		return
	}

	szlog.Debugf("tracer: request %s stage %q status %d after %s: %s%s",
		ctx.RequestId(), s.name, status, time.Since(begin),
		ctx.Host(), ctx.URI())
}
