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

package trace

import (
	"errors"
	"net/http"
	"time"

	"github.com/sentinez/core"
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/pkg/tracer"
	"github.com/sentinez/shared/rand"
	"github.com/sentinez/shared/zlog"
)

var _ corechains.ChainNode = (*stage)(nil)

func NewTracer(_ ...option.Option) corechains.ChainNode {
	return &Trace{
		Node: corechains.NewNode(),
	}
}

type Trace struct {
	*corechains.Node
}

// Handle makes sure the request has an id. The engine context assigns it
// once when the request is accepted; one is only generated here for a
// context that has none, so a request never gets a second id.
func (t *Trace) Handle(ctx corehttp.Context) error {
	if ctx.RequestId() == "" {
		ctx.SetRequestId(rand.NewXId(core.PrefixXRequestIdBytes))
	}

	return t.HandleNext(ctx)
}

// Wrap decorates node so that an error returned from it (and not already
// attributed to a deeper stage) is tagged with name and logged once.
// Because errors unwind from the innermost node outwards, the first
// stage to see an error is the one that produced it.
func Wrap(name string, node corechains.ChainNode) corechains.ChainNode {
	return &stage{name: name, node: node}
}

type stage struct {
	name string
	node corechains.ChainNode
}

func (s *stage) SetNext(next corechains.ChainNode) corechains.ChainNode {
	s.node.SetNext(next)
	return next
}

func (s *stage) Handle(ctx corehttp.Context) error {
	begin := time.Now()
	err := s.node.Handle(ctx)
	if err == nil {
		s.traceStatus(ctx, begin)
		return nil
	}

	var se *tracer.StageError
	if errors.As(err, &se) {
		return err
	}

	se = &tracer.StageError{Stage: s.name, Elapsed: time.Since(begin), Err: err}
	zlog.Errorf("tracer: request %s failed at stage %q status %d after %s: %v",
		ctx.RequestId(), se.Stage, ctx.StatusCode(), se.Elapsed, err)

	return se
}

// traceStatus logs a request answered with an error status without an
// error, e.g. a node that rendered a 403 page. Errors unwind from the
// innermost node outwards, so for a request the first stage logged is
// the one that answered; the stages after it only passed it through.
func (s *stage) traceStatus(ctx corehttp.Context, begin time.Time) {
	status := ctx.StatusCode()
	if status < http.StatusBadRequest {
		return
	}

	zlog.Debugf("tracer: request %s stage %q status %d after %s: %s%s",
		ctx.RequestId(), s.name, status, time.Since(begin),
		ctx.Host(), ctx.URI())
}
