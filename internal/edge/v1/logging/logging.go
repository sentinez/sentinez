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
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/types/net/http/v1"
	"github.com/sentinez/sentinez/internal/edge/engine/chains"
	httpxhz "github.com/sentinez/sentinez/pkg/network/httpx/hz"
	"github.com/sentinez/sentinez/pkg/zlog"
)

var _ chains.Handler = (*Logger)(nil)

func NewLogger() *Logger {
	return &Logger{
		BaseHandler: chains.New(),
		logger: zlog.NewJSONLogger(edgepb.GetMetaEdgeServiceKey(),
			common.LogKind_LOG_KIND_HTTP, zlog.LevelInfo,
		),
	}
}

type Logger struct {
	*chains.BaseHandler
	logger zlog.Logger
}

func (l *Logger) Handle(ctx *httpxhz.Context) error {
	zlog.Debugf("[edge][%s] >>> visit logger", ctx.GetReqID())

	requestResourceHost := string(ctx.Host())

	err := l.HandleNext(ctx)

	l.logger.Info("edge http request", &http.Log4HTTP{
		ReqId:         ctx.GetReqID(),
		Scheme:        string(ctx.URI().Scheme()),
		Host:          requestResourceHost,
		Path:          ctx.Path(),
		Method:        string(ctx.Method()),
		Status:        int32(ctx.Response.StatusCode()),
		RemoteAddress: ctx.RemoteAddr().String(),
		Protocol:      ctx.Request.Header.GetProtocol(),
		Query:         ctx.QueryArgs().String(),
		UserAgent:     string(ctx.UserAgent()),
	})

	return err
}
