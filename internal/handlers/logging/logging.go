// Copyright 2025 Duc-Hung Ho.
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
	"github.com/sentinez/core/szhttp/chains"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/dmz/dataplane/bpf"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/pkg/geoip"
	"github.com/sentinez/sentinez/pkg/pools/request"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/szlog"
)

// LoggingId is the chain stage name of Logging.
const LoggingId = "LOG"

// _headersSizeHint presizes the headers of an event: most requests fit.
const _headersSizeHint = 1 << 10

var (
	_ chains.Plugin                      = (*Logging)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*Logging)(nil)

	// _credentialHeaders are left out of the exported events.
	_credentialHeaders = [...][]byte{
		szbytestr.HeaderAuthorization,
		szbytestr.HeaderProxyAuth,
		szbytestr.HeaderCookie,
	}
)

// Logging logs every request of the chain and exports the records
// through the Env logger provider; without one they are not exported.
// The client location comes from the GeoLite2 databases compiled in.
type Logging struct {
	log szlog.LogCloser
	geo *geoip.DB
}

func (*Logging) ID() string { return LoggingId }

func (l *Logging) Provision(env *envplugin.Env) error {
	geo, err := env.GeoIP()
	if err != nil {
		return err
	}
	if names := geo.Loaded(); len(names) > 0 {
		szlog.Infof("edge: geoip databases: %v", names)
	}

	l.geo = geo
	l.log = env.Logger(typepb.LogType_LOG_TYPE_HTTP)
	return nil
}

func (l *Logging) Handle(ctx szhttp.Context, next chains.Handler) error {
	// szlog.Debug("[edge] >>> visit logging")
	ip := ctx.RequestIP()

	// the request as the client sent it: the next stages rewrite it, e.g.
	// the reverse proxy sets the upstream host
	var event *request.RequestEvent
	if l.log.V(szlog.LevelInfo.Int()) {
		event = newEvent(ctx, ip)
		l.geo.Fill(bytesconv.B2s(ip), event)
	}

	err := next.Handle(ctx)

	bw, _ := bpf.LookupBandwidth(bytesconv.B2s(ip))
	szlog.Infof("edge: lookup ip: %s bandwidth: %d", ip, bw)

	if event != nil {
		event.Status = int32(ctx.StatusCode())
		l.log.Info("http: request", event, event)
	}

	return err
}

func newEvent(ctx szhttp.Context, ip []byte) *request.RequestEvent {
	event := request.Acquire()

	event.Id = ctx.RequestId()
	event.Scheme = ctx.Scheme()
	event.Host = string(ctx.Host())
	event.Path = string(ctx.Path())
	event.Method = string(ctx.Method())
	event.Protocol = ctx.Protocol()
	event.UserAgent = string(ctx.Header(szbytestr.HeaderUserAgent))
	event.ContentType = string(ctx.Header(szbytestr.HeaderContentType))
	event.Queries = string(ctx.QueryStr())
	event.Headers = headersOf(ctx)
	event.RemoteAddress = string(ctx.RemoteAddr())
	event.ClientIp = string(ip)
	event.Fingerprint = ctx.JA4()
	event.Timestamp = ctx.RequestTime().UnixMilli()

	return event
}

// headersOf returns the request headers as "Name: value" lines, without
// the credentials ones.
func headersOf(ctx szhttp.Context) string {
	var b strings.Builder
	b.Grow(_headersSizeHint)

	ctx.VisitRequestHeaders(func(k, v []byte) {
		if isCredential(k) {
			return
		}
		b.Write(k)
		b.WriteString(": ")
		b.Write(v)
		b.WriteByte('\n')
	})

	return b.String()
}

func isCredential(k []byte) bool {
	for _, h := range _credentialHeaders {
		if bytes.EqualFold(k, h) {
			return true
		}
	}

	return false
}
