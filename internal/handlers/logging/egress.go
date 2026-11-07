// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"bytes"
	"strings"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/bytesconv"
	"github.com/sentinez/pkg/szlog"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/sentinez/pkg/pools/egress"
)

// _unknownLength the content length of a response without
// Content-Length, e.g. a streamed one.
const _unknownLength = -1

// proxied reports whether the egress of the request of x, its ctx.X(),
// is to be logged: the router set the upstream it proxied the request to,
// and the egress logger is enabled.
func (l *Logging) proxied(x *edgepb.ContextExtra) bool {
	return x.GetUpstream() != "" && l.egress.V(szlog.LevelInfo.Int())
}

// newEgressEvent returns the event of the leg from the reverse proxy to
// the upstream of ctx, recorded in x, ctx.X(), by the router: the request
// once proxied, and the answer of the origin.
func newEgressEvent(ctx szhttp.Context,
	x *edgepb.ContextExtra) *egress.EgressEvent {

	event := egress.Acquire()

	event.Id = ctx.RequestId()
	event.Upstream = x.GetUpstream()
	event.Method = string(ctx.Method())
	event.Path = string(ctx.Path())
	event.Queries = string(ctx.QueryStr())
	event.Protocol = ctx.Protocol()
	event.Scheme = schemeOf(x.GetUpstream())
	event.Status = int32(ctx.StatusCode())
	event.ContentLength = _unknownLength
	ctx.VisitResponseHeaders(func(k, v []byte) {
		setContentHeader(event, k, v)
	})
	event.LatencyMicros = x.GetUpstreamLatencyMicros()
	event.Timestamp = x.GetUpstreamTimestamp()
	event.ResourceId = ctx.ResourceId()
	event.TenantId = ctx.TenantId()

	return event
}

// setContentHeader sets the content type or length of event from the
// response header k, if it is one of them.
func setContentHeader(event *egress.EgressEvent, k, v []byte) {
	switch {
	case bytes.EqualFold(k, szbytestr.HeaderContentType):
		event.ContentType = string(v)
	case bytes.EqualFold(k, szbytestr.HeaderContentLength):
		event.ContentLength = contentLength(v)
	}
}

// contentLength parses the Content-Length v, _unknownLength when invalid.
func contentLength(v []byte) int64 {
	n, err := bytesconv.ParseUint(v)
	if err != nil {
		return _unknownLength
	}

	return int64(n)
}

// schemeOf the scheme of upstream, an URL, e.g. http.
func schemeOf(upstream string) string {
	scheme, _, _ := strings.Cut(upstream, "://")

	return scheme
}
