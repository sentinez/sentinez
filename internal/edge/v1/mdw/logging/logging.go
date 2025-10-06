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
	"github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/net/http/v1"
	"github.com/sentinez/sentinez/internal/edge/v1/chains"
	httpxhz "github.com/sentinez/sentinez/pkg/core/net/httpx/hz"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func NewLogger() *Logger {
	return &Logger{
		Base: &chains.Base{},
		logger: zlog.NewLoggingJSON(
			edge.GetMetaEdgeServiceKey(),
			common.LogKind_LOG_KIND_HTTP,
			zlog.LevelInfo,
		),
	}
}

type Logger struct {
	*chains.Base
	logger zlog.Logger
}

func (l *Logger) Handle(ctx *httpxhz.Context) error {
	zlog.Info("edge-handler: >>> Logger")

	requestResourceHost := string(ctx.Host())

	err := l.HandleNext(ctx)

	l.logger.Info("edge http request", &http.Log4HTTP{
		ReqId:         httpxhz.GetContextIdentify(ctx),
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
