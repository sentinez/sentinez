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
	"github.com/valyala/fasthttp"
)

func Writer(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	logger := zlog.NewLoggingJSON(
		edge.GetMetaEdgeServiceKey(),
		common.LogKind_LOG_KIND_HTTP,
		zlog.LevelInfo,
	)

	return func(ctx *fasthttp.RequestCtx) {
		requestResourceHost := string(ctx.Host())
		zlog.Debugf("[edge]: log resource host: %s", string(ctx.Host()))

		next(ctx)

		logger.Info("edge http request", &http.Log4HTTP{
			ReqId:         httpxf1.Identify(ctx),
			ReqScheme:     string(ctx.URI().Scheme()),
			ReqHost:       requestResourceHost,
			ReqPath:       string(ctx.Path()),
			ReqMethod:     string(ctx.Method()),
			RespStatus:    int32(ctx.Response.StatusCode()),
			ReqRemoteAddr: ctx.RemoteAddr().String(),
			ReqProtocol:   string(ctx.Request.Header.Protocol()),
			ReqQuery:      ctx.QueryArgs().String(),
			UserAgent:     string(ctx.UserAgent()),
		})
	}
}
