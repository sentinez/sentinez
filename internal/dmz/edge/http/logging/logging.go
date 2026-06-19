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
	corehttp "github.com/sentinez/core/http"
	corechains "github.com/sentinez/core/http/chains"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/dmz/edge/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/sentinez/internal/dmz/dataplane/ebpf"
	"github.com/sentinez/shared/zlog"
)

var _ corechains.ChainNode = (*Logger)(nil)

func NewLogger(logLevel zlog.Level) corechains.ChainNode {
	return &Logger{
		Node: corechains.NewNode(),
		logger: zlog.NewJSONLogger(edgepb.GetMetaEdgeServiceKey(),
			typepb.LogKind_LOG_KIND_HTTP, logLevel,
		),
	}
}

type Logger struct {
	*corechains.Node
	logger zlog.Logger
}

func (l *Logger) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit logger")

	requestResourceHost := string(ctx.Host())

	err := l.HandleNext(ctx)

	ip := ctx.RequestIP()
	bw, _ := ebpf.LookupBandwidth(ip)
	zlog.Infof("edge: lookup ip: %s bandwidth: %d", ip, bw)

	if l.logger.V(zlog.LevelInfo.Int()) {
		l.logger.Info("[http][request]", &typepb.RequestEvent{
			ReqId:         ctx.RequestId(),
			Scheme:        ctx.Scheme(),
			Host:          requestResourceHost,
			Path:          ctx.Path(),
			Method:        ctx.Method(),
			Status:        int32(ctx.StatusCode()),
			RemoteAddress: ctx.RemoteAddr(),
			Protocol:      ctx.Protocol(),
			Query:         ctx.QueryStr(),
			UserAgent:     ctx.Header(corehttp.HeaderUserAgent),
		})
	}

	return err
}
