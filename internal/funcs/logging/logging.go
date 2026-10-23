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
	"github.com/sentinez/core/szcmn/bytestr"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/dmz/dataplane/bpf"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/pkg/pools/request"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/szlog"
)

// LoggingId is the chain stage name of Logging.
const LoggingId = "LOG"

var (
	_ chains.Plugin                      = (*Logging)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*Logging)(nil)
)

// Logging logs every request of the chain and exports the records
// through the Env logger provider; without one they are not exported.
type Logging struct {
	log szlog.LogCloser
}

func (*Logging) ID() string { return LoggingId }

func (l *Logging) Provision(env *envplugin.Env) error {
	l.log = env.Logger(typepb.LogType_LOG_TYPE_HTTP)
	return nil
}

func (l *Logging) Handle(ctx szhttp.Context, next chains.Handler) error {
	// szlog.Debug("[edge] >>> visit logging")
	err := next.Handle(ctx)

	ip := ctx.RequestIP()
	bw, _ := bpf.LookupBandwidth(bytesconv.B2s(ip))
	szlog.Infof("edge: lookup ip: %s bandwidth: %d", ip, bw)

	event := request.Acquire()

	event.Id = ctx.RequestId()
	event.Scheme = ctx.Scheme()
	event.Host = string(ctx.Host())
	event.Path = string(ctx.Path())
	event.Method = string(ctx.Method())
	event.Status = int32(ctx.StatusCode())
	event.Protocol = ctx.Protocol()
	event.UserAgent = string(ctx.Header(bytestr.HeaderUserAgent))
	event.Queries = string(ctx.QueryStr())

	if l.log.V(szlog.LevelInfo.Int()) {
		l.log.Info("http: request", event, event)
	}

	return err
}
