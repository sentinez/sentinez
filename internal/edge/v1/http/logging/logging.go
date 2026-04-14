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
	"github.com/sentinez/core/chains"
	corehttp "github.com/sentinez/core/http"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"github.com/sentinez/shared/zlog"
)

var _ chains.Handler = (*Logger)(nil)

func NewLogger(logLevel zlog.Level) chains.Handler {
	return &Logger{
		BaseHandler: chains.New(),
		logger: zlog.NewJSONLogger(edgepb.GetMetaEdgeServiceKey(),
			typepb.LogKind_LOG_KIND_HTTP, logLevel,
		),
	}
}

type Logger struct {
	*chains.BaseHandler
	logger zlog.Logger
}

func (l *Logger) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit logger")

	requestResourceHost := string(ctx.Host())

	err := l.HandleNext(ctx)

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
