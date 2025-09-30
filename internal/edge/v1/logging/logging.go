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
	httpxf1 "github.com/sentinez/sentinez/pkg/core/net/httpx/f1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func WriterHandler(next httpxf1.RequestHandler) httpxf1.RequestHandler {
	logger := zlog.NewLoggingJSON(
		edge.GetMetaEdgeServiceKey(),
		common.LogKind_LOG_KIND_HTTP,
		zlog.LevelInfo,
	)

	return func(ctx *httpxf1.Context) error {
		requestResourceHost := string(ctx.Host())

		err := next(ctx)

		logger.Info("edge http request", &http.Log4HTTP{
			ReqId:         httpxf1.GetContextIdentify(ctx),
			Scheme:        string(ctx.URI().Scheme()),
			Host:          requestResourceHost,
			Path:          string(ctx.Path()),
			Method:        string(ctx.Method()),
			Status:        int32(ctx.Response.StatusCode()),
			RemoteAddress: ctx.RemoteAddr().String(),
			Protocol:      string(ctx.Request.Header.Protocol()),
			Query:         ctx.QueryArgs().String(),
			UserAgent:     string(ctx.UserAgent()),
		})

		return err
	}
}
