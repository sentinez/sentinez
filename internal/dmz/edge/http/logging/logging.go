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
	"github.com/sentinez/core/common/bytestr"
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/bpf"
	"github.com/sentinez/sentinez/internal/memory"
	"github.com/sentinez/sentinez/pkg/pools/request"
	"github.com/sentinez/sentinez/pkg/protocol"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/zlog"
)

var _ corechains.ChainNode = (*Logger)(nil)

func NewLogger(logLevel zlog.Level, _ *memory.MemStore) corechains.ChainNode {
	return &Logger{
		Node: corechains.NewNode(),
		log: zlog.NewLogCloser(edgepb.GetMetaEdgeServiceKey(),
			typepb.LogKind_LOG_KIND_HTTP, logLevel,
		),
	}
}

type Logger struct {
	*corechains.Node
	log zlog.LogCloser
}

func (l *Logger) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit logger")
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
	event.UserAgent = string(ctx.Header(bytestr.HeaderUserAgent))

	protocol.ParseQuery(ctx.Queries(), event.Queries)
	protocol.ParseHeader(ctx.Headers(), event.Headers)

	if l.log.V(zlog.LevelInfo.Int()) {
		l.log.Info("http: request", event, event)
	}

	return err
}
