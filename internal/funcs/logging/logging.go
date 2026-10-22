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
	"github.com/sentinez/core/szcmn/szbytestr"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/szchains"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/dmz/dataplane/bpf"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/pkg/pools/request"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/zlog"
)

var _ szchains.ChainNode = (*Logging)(nil)

// NewLogging logs every request of the chain and exports the records
// through the option.WithLogger provider; without one they are not
// exported.
func NewLogging(opts ...option.Option) szchains.ChainNode {
	o := option.New(opts...)

	return &Logging{
		Node: szchains.NewNode(),
		log: zlog.NewLogCloser(
			o.Named,
			typepb.LogType_LOG_TYPE_HTTP,
			o.Level,
			o.OTLP(),
		),
	}
}

type Logging struct {
	*szchains.Node
	log zlog.LogCloser
}

func (l *Logging) Handle(ctx szhttp.Context) error {
	// zlog.Debug("[edge] >>> visit logging")
	err := l.HandleNext(ctx)

	ip := ctx.RequestIP()
	bw, _ := bpf.LookupBandwidth(bytesconv.B2s(ip))
	zlog.Infof("edge: lookup ip: %s bandwidth: %d", ip, bw)

	event := request.Acquire()

	event.Id = ctx.RequestId()
	event.Scheme = ctx.Scheme()
	event.Host = string(ctx.Host())
	event.Path = string(ctx.Path())
	event.Method = string(ctx.Method())
	event.Status = int32(ctx.StatusCode())
	event.Protocol = ctx.Protocol()
	event.UserAgent = string(ctx.Header(szbytestr.HeaderUserAgent))
	event.Queries = string(ctx.QueryStr())

	if l.log.V(zlog.LevelInfo.Int()) {
		l.log.Info("http: request", event, event)
	}

	return err
}
