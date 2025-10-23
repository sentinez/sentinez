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

package httpxhz

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sentinez/sentinez"
	"github.com/sentinez/sentinez/pkg/htmlx"
	"github.com/sentinez/sentinez/pkg/x/uuidx"
)

const (
	HeaderXRequest = "X-Request-ID"
)

func WrapHandler(next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {

		c.Request.Header.Add(HeaderXRequest,
			uuidx.NewIDHex(sentinez.PrefixRequestID))

		ctx = setRequestTime(ctx)

		next(ctx, c)
	}
}

func Forbidden(ctx *Context) error {
	err := ctx.Render(http.StatusForbidden, htmlx.Forbidden(ctx.GetReqID()))
	if err != nil {
		ctx.Response.ResetBody()
		return ctx.String(http.StatusForbidden, "Access denied")
	}

	return nil
}

func InternalServerError(ctx *Context) error {
	err := ctx.Render(
		http.StatusInternalServerError, htmlx.InternalError(ctx.GetReqID()))
	if err != nil {
		ctx.Response.ResetBody()
		return ctx.String(
			http.StatusInternalServerError, "Internal server error")
	}

	return nil
}

func NotFound(ctx *Context) error {
	err := ctx.Render(http.StatusNotFound, htmlx.NotFound(ctx.GetReqID()))
	if err != nil {
		ctx.Response.ResetBody()
		return ctx.String(http.StatusNotFound, "Not found")
	}

	return nil
}
